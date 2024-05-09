package api

import (
	"context"
	sparkpb "github.com/zcubbs/spark/gen/proto/go/spark/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) DeleteJob(ctx context.Context, req *sparkpb.DeleteJobRequest) (*sparkpb.DeleteJobResponse, error) {
	err := s.k8sRunner.Delete(ctx, req.JobId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create job: %v", err)
	}

	return &sparkpb.DeleteJobResponse{}, nil
}
