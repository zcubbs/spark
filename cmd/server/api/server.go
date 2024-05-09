package api

import (
	"context"
	"crypto/tls"
	"embed"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/zcubbs/spark/cmd/server/config"
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
)

type Server struct {
	sparkPb.UnimplementedSparkServiceServer

	cfg       *config.Configuration
	embedOpts []EmbedAssetsOpts
	k8sRunner *k8sJobs.Runner

	limiter *rate.Limiter
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

func (s *Server) StartGrpcServer() {
	var tlsOpt grpc.ServerOption
	if s.cfg.GrpcServer.Tls.Enabled {
		var err error
		tlsOpt, err = newServerTlsOptions(s.cfg.GrpcServer)
		if err != nil {
			log.Fatal("cannot create new server tls options", "error", err)
		}
	} else {
		tlsOpt = grpc.EmptyServerOption{}
	}

	// Logging and rate limiting interceptors
	unifiedInterceptors := grpc.ChainUnaryInterceptor(logger.GrpcLogger, s.unaryRateLimitInterceptor)

	grpcServer := grpc.NewServer(unifiedInterceptors, tlsOpt)
	sparkPb.RegisterSparkServiceServer(grpcServer, s)

	if s.cfg.GrpcServer.EnableReflection {
		reflection.Register(grpcServer)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.GrpcServer.Port))
	if err != nil {
		log.Fatal("cannot listen", "error", err, "port", s.cfg.GrpcServer.Port)
	}

	log.Info("🟢 starting grpc server", "port", s.cfg.GrpcServer.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("cannot start grpc server", "error", err)
	}
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

	// server options
	httpSrv := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.cfg.HttpServer.Port),
		ReadHeaderTimeout: s.cfg.HttpServer.ReadHeaderTimeout,
		Handler:           handler,
	}

	log.Info("🟢 starting HTTP Gateway server", "port", s.cfg.HttpServer.Port)
	if err := httpSrv.ListenAndServe(); err != nil {
		log.Fatal("cannot start http server", "error", err)
	}
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
	log.Info("serving embedded assets", "path", opts.Path)
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
