package api

import (
	"context"
	sparkpb "github.com/zcubbs/spark/gen/proto/go/spark/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetJobLogs(_ context.Context, req *sparkpb.GetJobLogsRequest) (*sparkpb.GetJobLogsResponse, error) {
	// Get the job logs
	logs, err := s.k8sRunner.GetLogsForTaskFromDB(req.JobId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get logs: %v", err)
	}

	// Return the logs
	return &sparkpb.GetJobLogsResponse{
		Logs: logs,
	}, nil
}
