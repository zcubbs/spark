package api

import (
	"context"
	"crypto/rand"
	"fmt"
	sparkpb "github.com/zcubbs/spark/gen/proto/go/spark/v1"
	k8sJobs "github.com/zcubbs/spark/pkg/k8s/jobs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"regexp"
)

func (s *Server) QueueJob(_ context.Context, req *sparkpb.QueueJobRequest) (*sparkpb.QueueJobResponse, error) {
	var id string
	if req.JobId == "" {
		var err error
		id, err = generateRandomJobId()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to generate random job id: %v", err)
		}
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

// generateRandomJobId generates a secure and random job ID
func generateRandomJobId() (string, error) {
	var charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 15) // Length of the random part

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	var result = "spark-job-"
	for _, byteVal := range b {
		result += string(charset[byteVal%byte(len(charset))])
	}

	return result, nil
}
