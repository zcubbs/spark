package k8sJobs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/charmbracelet/log"
	"io"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/utils/ptr"
)

func (r *Runner) createAndMonitorJob(ctx context.Context, namespace string, task Task) (*batchv1.Job, error) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      task.ID,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    task.ID,
							Image:   task.Image,
							Command: task.Command,
						},
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
			ActiveDeadlineSeconds: ptr.To(int64(task.Timeout)), // Ensure jobs respect the task timeout
		},
	}

	log.Debug("Creating job", "jobId", task.ID, "image", task.Image, "command", task.Command)
	job, err := r.cs.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		log.Error("Failed to create job", "error", err)
		return nil, err
	}

	log.Debug("Waiting for job to complete...", "jobId", task.ID, "image", task.Image, "command", task.Command)
	if err = r.waitForJobCompletion(ctx, job); err != nil {
		log.Error("Failed to monitor job", "error", err)
		return nil, fmt.Errorf("job monitoring failed: %v", err)
	}

	return job, nil
}

func (r *Runner) waitForJobCompletion(ctx context.Context, job *batchv1.Job) error {
	watcher, err := r.cs.BatchV1().Jobs(job.Namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", job.Name),
	})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("job watch channel closed")
			}
			switch event.Type {
			case watch.Error:
				return fmt.Errorf("error watching job: %v", event.Object)
			case watch.Deleted:
				return fmt.Errorf("job deleted unexpectedly")
			case watch.Added, watch.Modified:
				job, ok := event.Object.(*batchv1.Job)
				if !ok {
					return fmt.Errorf("unexpected object type")
				}
				if done, err := r.evaluateJobStatus(job); done {
					return err
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (r *Runner) evaluateJobStatus(job *batchv1.Job) (bool, error) {
	for _, condition := range job.Status.Conditions {
		switch condition.Type {
		case batchv1.JobComplete:
			if condition.Status == corev1.ConditionTrue {
				return true, nil
			}
		case batchv1.JobFailed:
			if condition.Status == corev1.ConditionTrue {
				return true, fmt.Errorf("job failed: %s", condition.Message)
			}
		}
	}
	return false, nil // Job has not yet reached a definitive state
}

// Delete deletes a pod in a Kubernetes cluster.
func (r *Runner) delete(ctx context.Context, jobId string) {
	deletePolicy := metav1.DeletePropagationForeground
	log.Debug("Deleting job", "jobId", jobId)
	err := r.cs.BatchV1().Jobs(r.namespace).Delete(ctx, jobId, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		log.Error("Failed to delete job", "error", err, "jobId", jobId)
	}
}

// GetLogs gets the logs of a running pod in a Kubernetes cluster.
func (r *Runner) getLogs(ctx context.Context, jobId string) (string, error) {
	pods, err := r.cs.CoreV1().Pods(r.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobId),
	})
	if err != nil {
		return "", fmt.Errorf("error fetching pods: %v", err)
	}

	var logsAggregate string
	for _, pod := range pods.Items {
		logOpts := &corev1.PodLogOptions{}
		req := r.cs.CoreV1().Pods(r.namespace).GetLogs(pod.Name, logOpts)
		logs, err := req.Stream(ctx)
		if err != nil {
			// Log the error and continue with the next pod
			log.Errorf("error getting logs for pod %s: %v", pod.Name, err)
			continue
		}

		// Handle resource closing and error checking properly within the loop
		logsData, err := readAndCloseStream(logs)
		if err != nil {
			log.Errorf("error reading logs for pod %s: %v", pod.Name, err)
			continue
		}

		if logsAggregate != "" && logsData != "" {
			logsAggregate += "\n"
		}
		logsAggregate += logsData
	}

	return logsAggregate, nil
}

// readAndCloseStream reads data from a ReadCloser, closes it, and returns the data
func readAndCloseStream(rc io.ReadCloser) (string, error) {
	defer func(rc io.ReadCloser) {
		err := rc.Close()
		if err != nil {
			log.Error("Failed to close stream", "error", err)
		}
	}(rc) // defer is called after the surrounding function returns (readAndCloseStream in this case)
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(rc)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
