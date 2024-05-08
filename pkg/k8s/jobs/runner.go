package k8sJobs

import (
	"bytes"
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Options is a struct that contains the options for running a Kubernetes pod.
type Options struct {
	// PodName is the name of the pod.
	Name string
	// Namespace is the namespace of the pod.
	Namespace string
	// Image is the image of the pod.
	Image string
	// Command is the command to run in the pod.
	Command []string
	// KubeconfigPath is the path to the kubeconfig file.
	KubeconfigPath string
}

// Runner is an interface that defines the methods for running a Kubernetes pod.
type Runner struct {
	cs *kubernetes.Clientset
}

func New(kubeconfig string) (*Runner, error) {
	var (
		kConfig *rest.Config
		cs      *kubernetes.Clientset
		err     error
	)

	if kubeconfig == "" {

		kConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load in-cluster config: %v", err)
		}
	} else {
		fmt.Println("Using kubeconfig file")
		kConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig file: %v", err)
		}
	}

	cs, err = kubernetes.NewForConfig(kConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	return &Runner{
		cs: cs,
	}, nil
}

// Run runs a pod in a Kubernetes cluster.
func (r *Runner) Run(ctx context.Context, options Options) (*batchv1.Job, error) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.Name,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    options.Name,
							Image:   options.Image,
							Command: options.Command,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit: int32Ptr(0), // No retries
		},
	}
	return r.cs.BatchV1().Jobs(options.Namespace).Create(ctx, job, metav1.CreateOptions{})
}

// Kill stops a running pod in a Kubernetes cluster.
func (r *Runner) Kill() error {
	// Implement me
	return nil
}

// Delete deletes a pod in a Kubernetes cluster.
func (r *Runner) Delete(ctx context.Context, options Options) error {
	deletePolicy := metav1.DeletePropagationForeground
	err := r.cs.BatchV1().Jobs(options.Namespace).Delete(ctx, options.Name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to delete job: %v", err)
	}

	return nil
}

// GetLogs gets the logs of a running pod in a Kubernetes cluster.
func (r *Runner) GetLogs(ctx context.Context, options Options) (string, error) {
	pods, err := r.cs.CoreV1().Pods(options.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", options.Name),
	})
	if err != nil {
		return "", fmt.Errorf("error fetching pods: %v", err)
	}

	var logsAggregate string
	for _, pod := range pods.Items {
		logOpts := &corev1.PodLogOptions{}
		req := r.cs.CoreV1().Pods(options.Namespace).GetLogs(pod.Name, logOpts)
		logs, err := req.Stream(ctx)
		if err != nil {
			continue // Log error or handle it as needed
		}
		defer logs.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(logs)
		logsAggregate += buf.String() + "\n" // Collect logs for all pods
	}

	return logsAggregate, nil
}

// WatchStatus watches the status of a running pod in a Kubernetes cluster.
func (r *Runner) WatchStatus() error {
	// Implement me
	return nil
}

func LoadK8sConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig == "" {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func int32Ptr(i int32) *int32 { return &i }
