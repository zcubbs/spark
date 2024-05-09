package k8sJobs

import (
	"context"
	"fmt"
	"github.com/tidwall/buntdb"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path"
	"path/filepath"
	"sync"
)

// Runner manages the lifecycle of Kubernetes jobs.
type Runner struct {
	cs                *kubernetes.Clientset
	maxConcurrentJobs int
	wg                sync.WaitGroup
	taskChan          chan Task
	quit              chan struct{}
	namespace         string

	db *buntdb.DB
}

func New(ctx context.Context, kubeconfig string, maxConcurrentJobs, queueBufferSize int) (*Runner, error) {
	kConfig, err := loadK8sConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load k8s config: %v", err)
	}

	cs, err := kubernetes.NewForConfig(kConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s clientset: %v", err)
	}

	ns := determineNamespace()

	db, err := buntdb.Open("spark.db") // Open the data store
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	r := &Runner{
		cs:                cs,
		maxConcurrentJobs: maxConcurrentJobs,
		taskChan:          make(chan Task, queueBufferSize),
		quit:              make(chan struct{}),
		namespace:         ns,
		db:                db,
	}

	go r.processTasks(ctx)

	return r, nil
}

// Shutdown stops the runner.
func (r *Runner) Shutdown() error {
	close(r.quit)
	r.wg.Wait()
	return r.db.Close()
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
