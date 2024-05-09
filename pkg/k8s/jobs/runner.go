package k8sJobs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/charmbracelet/log"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path"
	"path/filepath"
	"sync"
)

type Task struct {
	ID      string
	Command []string
	Image   string
}

// Runner manages the lifecycle of Kubernetes jobs.
type Runner struct {
	cs                *kubernetes.Clientset
	maxConcurrentJobs int
	wg                sync.WaitGroup
	taskChan          chan Task
	quit              chan struct{}
	namespace         string
}

func New(ctx context.Context, kubeconfig string, maxConcurrentJobs int) (*Runner, error) {
	kConfig, err := loadK8sConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load k8s config: %v", err)
	}

	cs, err := kubernetes.NewForConfig(kConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s clientset: %v", err)
	}

	ns := determineNamespace()

	r := &Runner{
		cs:                cs,
		maxConcurrentJobs: maxConcurrentJobs,
		taskChan:          make(chan Task, 100), // Buffer can be adjusted based on expected task load
		quit:              make(chan struct{}),
		namespace:         ns,
	}

	go r.processTasks(ctx)

	return r, nil
}

func (r *Runner) processTasks(ctx context.Context) {
	for {
		select {
		case task := <-r.taskChan:
			r.wg.Add(1)
			go func(t Task) {
				defer r.wg.Done()
				// Process the task
				_, err := r.createAndMonitorJob(ctx, r.namespace, t)
				if err != nil {
					log.Error("Failed to create and monitor job", "error", err)
				}
			}(task)
		case <-r.quit:
			return
		}
	}
}

// AddTask adds a task to the runner.
func (r *Runner) AddTask(t Task) error {
	select {
	case r.taskChan <- t:
		return nil // Successfully added the task
	default:
		// Handle the case when taskChan is full
		return fmt.Errorf("task queue is full")
	}
}

// Shutdown stops the runner.
func (r *Runner) Shutdown() {
	close(r.quit)
	r.wg.Wait()
}

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
		},
	}

	// Create the Kubernetes job
	job, err := r.cs.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		log.Error("Failed to create job", "error", err)
		return nil, err
	}

	// Monitor the job status until completion or failure
	err = r.waitForJobCompletion(ctx, job)
	if err != nil {
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

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added, watch.Modified:
			job, ok := event.Object.(*batchv1.Job)
			if !ok {
				return fmt.Errorf("unexpected type")
			}

			for _, condition := range job.Status.Conditions {
				if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
					return nil
				} else if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
					return fmt.Errorf("job failed")
				}
			}
		case watch.Deleted:
			return fmt.Errorf("job deleted")
		}
	}

	return nil
}

// Kill stops a running pod in a Kubernetes cluster.
func (r *Runner) Kill() error {
	// Implement me
	return nil
}

// Delete deletes a pod in a Kubernetes cluster.
func (r *Runner) Delete(ctx context.Context, jobId string) error {
	deletePolicy := metav1.DeletePropagationForeground
	err := r.cs.BatchV1().Jobs(r.namespace).Delete(ctx, jobId, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to delete job: %v", err)
	}

	return nil
}

// GetLogs gets the logs of a running pod in a Kubernetes cluster.
func (r *Runner) GetLogs(ctx context.Context, jobId string) (string, error) {
	pods, err := r.cs.CoreV1().Pods(r.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf(jobId),
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
			continue // Log error or handle it as needed
		}
		defer logs.Close()
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(logs)
		if err != nil {
			return "", err
		}
		logsAggregate += buf.String() + "\n" // Collect logs for all pods
	}

	return logsAggregate, nil
}

// WatchStatus watches the status of a running pod in a Kubernetes cluster.
func (r *Runner) WatchStatus() error {
	// Implement me
	return nil
}

// loadK8sConfig loads the Kubernetes client configuration.
func loadK8sConfig(kubeconfig string) (*rest.Config, error) {
	var (
		kConfig *rest.Config
		err     error
	)
	if kubeconfig == "" {
		kConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load in-cluster config: %v", err)
		}
	} else {
		fmt.Println("Using kubeconfig file")

		if kubeconfig == "default" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %v", err)
			}

			kubeconfig = path.Clean(
				filepath.Join(homeDir, ".kube", "config"),
			)

			if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
				return nil, fmt.Errorf("kubeconfig file does not exist: %v", err)
			}
		}

		kConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig file: %v", err)
		}
	}

	return kConfig, nil
}

// determineNamespace retrieves the namespace the application is running under
func determineNamespace() string {
	namespace, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "default" // Default to "default" namespace if running locally
	}
	return string(namespace)
}

func int32Ptr(i int32) *int32 { return &i }

// GetNamespace returns the namespace of the runner.
func (r *Runner) GetNamespace() string {
	return r.namespace
}
