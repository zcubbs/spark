package api

import (
	"context"
	sparkpb "github.com/zcubbs/spark/gen/proto/go/spark/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetJobStatus(_ context.Context, req *sparkpb.GetJobStatusRequest) (*sparkpb.GetJobStatusResponse, error) {
	// Get the job status
	st, err := s.k8sRunner.GetStatusForTaskFromDB(req.JobId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get job status: %v", err)
	}

	// Return the status
	return &sparkpb.GetJobStatusResponse{
		Status: st,
	}, nil
}
