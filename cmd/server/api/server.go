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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
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
}

func NewServer(cfg *config.Configuration) (*Server, error) {
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
	}

	return s, nil
}

func (s *Server) StartGrpcServer() {
	grpcLogger := grpc.UnaryInterceptor(logger.GrpcLogger)

	var tlsOpt grpc.ServerOption
	if s.cfg.GrpcServer.Tls.Enabled {
		var err error
		tlsOpt, err = newServerTlsOptions(s.cfg.GrpcServer)
		if err != nil {
			log.Fatal("cannot create new server tls options", "error", err)
		}
	} else {
		log.Warn("🔴 grpc server is running without TLS")
		tlsOpt = grpc.EmptyServerOption{}
	}

	grpcServer := grpc.NewServer(grpcLogger, tlsOpt)
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

func (s *Server) StartHttpGateway() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	grpcMux := newGrpcRuntimeServerMux()

	err := sparkPb.RegisterSparkServiceHandlerServer(ctx, grpcMux, s)
	if err != nil {
		log.Fatal("cannot register handler server", "error", err)
	}

	mux := http.NewServeMux()

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
	handler := logger.HttpLogger(mux)

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
