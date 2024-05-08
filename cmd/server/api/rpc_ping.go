package api

import (
	"context"
	sparkpb "github.com/zcubbs/spark/gen/proto/go/spark/v1"
)

func (s *Server) Ping(_ context.Context, _ *sparkpb.PingRequest) (*sparkpb.PingResponse, error) {
	return &sparkpb.PingResponse{
		Message:   "Pong",
		Version:   s.cfg.Version,
		Commit:    s.cfg.Commit,
		BuildTime: s.cfg.Date,
	}, nil
}
