package api

import (
	"context"
	"fmt"
	sparkpb "github.com/zcubbs/spark/gen/proto/go/spark/v1"
	k8sJobs "github.com/zcubbs/spark/pkg/k8s/jobs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"regexp"
)

func (s *Server) QueueJob(_ context.Context, req *sparkpb.QueueJobRequest) (*sparkpb.QueueJobResponse, error) {
	var id string
	if req.JobId == "" {
		id = generateRandomJobId()
	} else {
		if err := validateId(req.JobId); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid [id]: %v", err)
		}
		id = req.JobId
	}

	// Add task to the runner
	t := k8sJobs.Task{
		ID:      id,
		Image:   req.Image,
		Command: req.Command,
	}
	err := s.k8sRunner.AddTask(t)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add task: %v", err)
	}

	return &sparkpb.QueueJobResponse{
		Id:        id,
		Namespace: s.k8sRunner.GetNamespace(),
	}, nil
}

// Validate Id
// Must only contain lowercase alphanumeric characters, dashes, and numbers
// Must be between 1 and 25 characters long
// Must not start or end with a dash
// Must not contain two consecutive dashes
// Must not contain spaces
func validateId(id string) error {
	if len(id) < 1 || len(id) > 25 {
		return fmt.Errorf("id must be between 1 and 25 characters long")
	}
	if !regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`).MatchString(id) {
		return fmt.Errorf("id must only contain lowercase alphanumeric characters, dashes, and numbers")
	}
	return nil
}

// generateRandomJobId generates a random job id
func generateRandomJobId() string {
	// implement this function
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, 15)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return "spark-job-" + string(b)
}
