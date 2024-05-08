package api

import (
	"context"
	sparkpb "github.com/zcubbs/spark/gen/proto/go/spark/v1"
	k8sJobs "github.com/zcubbs/spark/pkg/k8s/jobs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func (s *Server) RunJob(ctx context.Context, req *sparkpb.RunJobRequest) (*sparkpb.RunJobResponse, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(req.Timeout))
	defer cancel()

	job, err := s.k8sRunner.Run(timeoutCtx, k8sJobs.Options{
		Name:           req.Name,
		Namespace:      req.Namespace,
		Image:          req.Image,
		Command:        []string{req.Command},
		KubeconfigPath: s.cfg.KubeconfigPath,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create job: %v", err)
	}

	return &sparkpb.RunJobResponse{
		Id:        job.Name,
		Name:      job.Name,
		Namespace: job.Namespace,
		Image:     job.Spec.Template.Spec.Containers[0].Image,
		Command:   job.Spec.Template.Spec.Containers[0].Command[0],
		Status:    job.Status.String(),
		Created:   job.CreationTimestamp.String(),
		Started:   job.Status.StartTime.String(),
		Finished:  job.Status.CompletionTime.String(),
		Logs:      job.Status.Conditions[0].Message,
	}, nil
}
