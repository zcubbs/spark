package api

import (
	"context"
	"crypto/tls"
	"embed"
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/zcubbs/spark/cmd/server/config"
	"github.com/zcubbs/spark/cmd/server/web"
	"github.com/zcubbs/spark/gen/openapi"
	sparkPb "github.com/zcubbs/spark/gen/proto/go/spark/v1"
	"github.com/zcubbs/spark/internal/logger"
	k8sJobs "github.com/zcubbs/spark/pkg/k8s/jobs"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"time"
)

type Server struct {
	sparkPb.UnimplementedSparkServiceServer

	cfg       *config.Configuration
	embedOpts []EmbedAssetsOpts
	k8sRunner *k8sJobs.Runner

	limiter *rate.Limiter

	grpcServer  *grpc.Server
	httpGateway *http.Server
	webServer   *http.Server
}

func NewServer(cfg *config.Configuration, jobRunner *k8sJobs.Runner) (*Server, error) {
	var embeds []EmbedAssetsOpts
	swaggerEmbed := EmbedAssetsOpts{
		Dir:    openapi.OpenApiFs,
		Path:   "/swagger/",
		Prefix: ".",
	}
	embeds = append(embeds, swaggerEmbed)

	s := &Server{
		cfg:       cfg,
		embedOpts: embeds,
		k8sRunner: jobRunner,
		limiter:   rate.NewLimiter(rate.Limit(cfg.RateLimitRequestsPerSecond), cfg.RateLimitBurst),
	}

	return s, nil
}

func (s *Server) StartGrpcServer(ctx context.Context) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.GrpcServer.Port))
	if err != nil {
		log.Fatal("cannot listen", "error", err, "port", s.cfg.GrpcServer.Port)
	}

	var tlsOpt grpc.ServerOption
	if s.cfg.GrpcServer.Tls.Enabled {
		tlsOpt, err = newServerTlsOptions(s.cfg.GrpcServer)
		if err != nil {
			log.Fatal("cannot create new server tls options", "error", err)
		}
	} else {
		tlsOpt = grpc.EmptyServerOption{}
	}

	// Logging and rate limiting interceptors
	unifiedInterceptors := grpc.ChainUnaryInterceptor(logger.GrpcLogger, s.unaryRateLimitInterceptor)

	s.grpcServer = grpc.NewServer(unifiedInterceptors, tlsOpt)
	sparkPb.RegisterSparkServiceServer(s.grpcServer, s)

	if s.cfg.GrpcServer.EnableReflection {
		reflection.Register(s.grpcServer)
	}

	go func() {
		log.Info("🟢 starting grpc server", "port", s.cfg.GrpcServer.Port)
		if err := s.grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Error("failed to serve gRPC", "error", err)
		}
	}()

	// Listen for context cancellation and gracefully stop the server
	go func() {
		<-ctx.Done()
		s.grpcServer.GracefulStop()
		log.Info("gRPC server shutdown gracefully")
	}()
}

func (s *Server) StartHttpGateway(ctx context.Context) {
	mux := http.NewServeMux()
	grpcMux := newGrpcRuntimeServerMux()
	err := sparkPb.RegisterSparkServiceHandlerServer(ctx, grpcMux, s)
	if err != nil {
		log.Fatal("cannot register handler server", "error", err)
	}

	// add embedded assets handler
	err = mime.AddExtensionType(".svg", "image/svg+xml")
	if err != nil {
		log.Fatal("couldn't add extension type", "err", err.Error())
	}
	for _, opts := range s.embedOpts {
		mux.Handle(opts.Path, newFileServerHandler(opts))
	}

	// add grpc handler
	mux.Handle("/", grpcMux)

	// add rate limit middleware
	rateLimitMux := rateLimitMiddleware(s.limiter, mux)

	// add logger middleware
	handler := logger.HttpLogger(rateLimitMux)

	// Cors
	origins := handlers.AllowedOrigins([]string{"*"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	headers := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	handler = handlers.CORS(origins, methods, headers)(handler)

	s.httpGateway = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.cfg.HttpServer.ApiPort),
		ReadHeaderTimeout: s.cfg.HttpServer.ReadHeaderTimeout,
		Handler:           handler,
	}

	go func() {
		log.Info("🟢 starting HTTP Gateway", "port", s.cfg.HttpServer.ApiPort)
		if err := s.httpGateway.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP gateway server stopped unexpectedly", "error", err)
		}
	}()

	// Listen for context cancellation and shutdown the HTTP gateway server gracefully
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpGateway.Shutdown(shutdownCtx); err != nil {
			log.Error("failed to shutdown HTTP gateway", "error", err)
		}
		log.Info("HTTP gateway shutdown gracefully")
	}()
}

func (s *Server) StartWebServer(ctx context.Context) {
	mux := http.NewServeMux()

	h, err := web.NewHandler(s.k8sRunner)
	if err != nil {
		log.Fatal("cannot create new handler", "error", err)
	}
	h.RegisterRoutes(mux)

	s.webServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.cfg.HttpServer.WebPort),
		ReadHeaderTimeout: s.cfg.HttpServer.ReadHeaderTimeout,
		Handler:           mux,
	}

	go func() {
		log.Info("🟢 starting web server", "port", s.cfg.HttpServer.WebPort)
		if err := s.webServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Web server stopped unexpectedly", "error", err)
		}
	}()

	// Listen for context cancellation and shutdown the web server gracefully
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.webServer.Shutdown(shutdownCtx); err != nil {
			log.Error("failed to shutdown web server", "error", err)
		}
		log.Info("Web server shutdown gracefully")
	}()
}

func newGrpcRuntimeServerMux() *runtime.ServeMux {
	jsonOpts := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	return runtime.NewServeMux(jsonOpts)
}

func newServerTlsOptions(cfg config.GrpcServerConfig) (grpc.ServerOption, error) {
	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(cfg.Tls.Cert, cfg.Tls.Key)
	if err != nil {
		return nil, fmt.Errorf("could not load server key pair: %w", err)
	}

	// Create the TLS credentials
	return grpc.Creds(credentials.NewServerTLSFromCert(&certificate)), nil
}

type EmbedAssetsOpts struct {
	// The directory to embed.
	Dir    embed.FS
	Path   string
	Prefix string
}

func newFileServerHandler(opts EmbedAssetsOpts) http.Handler {
	log.Debug("serving embedded assets", "path", opts.Path)
	sub, err := fs.Sub(opts.Dir, opts.Prefix)
	if err != nil {
		log.Fatal("cannot serve embedded assets", "error", err)
	}
	dir := http.FileServer(http.FS(sub))

	return http.StripPrefix(opts.Path, dir)
}

func (s *Server) unaryRateLimitInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if !s.limiter.Allow() {
		return nil, status.Errorf(codes.ResourceExhausted, "request limit exceeded")
	}
	return handler(ctx, req)
}

func rateLimitMiddleware(limiter *rate.Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Shutdown() {
	// Context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shut down HTTP Gateway
	if err := s.httpGateway.Shutdown(ctx); err != nil {
		log.Error("failed to shutdown HTTP gateway", "error", err)
	}

	// Shut down Web Server
	if err := s.webServer.Shutdown(ctx); err != nil {
		log.Error("failed to shutdown web server", "error", err)
	}

	// Stop gRPC server
	s.grpcServer.GracefulStop()
	log.Info("Servers have been shutdown gracefully")
}
