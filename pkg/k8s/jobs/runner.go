package k8sJobs

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
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
	// Timeout is the timeout for running the pod.
	Timeout int
	// KubeconfigPath is the path to the kubeconfig file.
	KubeconfigPath string
}

// Runner is an interface that defines the methods for running a Kubernetes pod.
type Runner struct {
	options Options
	cs      *kubernetes.Clientset
}

func New(options Options) *Runner {
	return &Runner{
		options: options,
	}
}

// Run runs a pod in a Kubernetes cluster.
func (r *Runner) Run() (*batchv1.Job, error) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.options.Name,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    r.options.Name,
							Image:   r.options.Image,
							Command: r.options.Command,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit: int32Ptr(0), // No retries
		},
	}
	return r.cs.BatchV1().Jobs(r.options.Namespace).Create(context.TODO(), job, metav1.CreateOptions{})
}

// Kill stops a running pod in a Kubernetes cluster.
func (r *Runner) Kill() error {
	// Implement me
	return nil
}

// GetLogs gets the logs of a running pod in a Kubernetes cluster.
func (r *Runner) GetLogs() (string, error) {
	// Implement me
	return "", nil
}

// WatchStatus watches the status of a running pod in a Kubernetes cluster.
func (r *Runner) WatchStatus() error {
	// Implement me
	return nil
}

func int32Ptr(i int32) *int32 { return &i }

// loadConfig loads the Kubernetes client configuration.
func (r *Runner) loadConfig(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath == "" {
		fmt.Println("Using in-cluster config")
		return rest.InClusterConfig()
	}
	fmt.Println("Using kubeconfig file")
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

// getNamespace retrieves the namespace the application is running under.
func (r *Runner) getNamespace() string {
	namespace, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "default" // Default to "default" namespace if running locally
	}
	return string(namespace)
}
