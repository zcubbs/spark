# k8sJobs Package

The `k8sJobs` package provides a simple and efficient way to manage and execute Kubernetes jobs in Go. It uses channels to queue tasks and control their execution, allowing for concurrent job processing within specified limits.

## Overview

This package is designed to help developers integrate Kubernetes job management into their applications with minimal setup. It handles the creation, execution, and monitoring of Kubernetes jobs based on tasks that are queued through a channel.

## Key Components

- **Task:** Represents a Kubernetes job with a specific command, image, and identifiers.
- **Runner:** Manages tasks, orchestrates the creation and monitoring of Kubernetes jobs.

## Queuing System

The queuing system in `k8sJobs` is implemented using Go channels, which serve as a robust mechanism for managing concurrency and task execution order. This approach ensures that tasks are executed in a controlled manner, respecting the `maxConcurrentJobs` limit set during initialization.

### Task Channel

Tasks are submitted to a `taskChan`, a buffered channel where tasks wait until they can be processed. The size of the buffer defines the maximum number of tasks that can be pending execution, which is determined by the `maxConcurrentJobs` parameter.

### Runner Process

The `Runner` continuously listens for new tasks on the `taskChan`. Each task is processed by launching a separate goroutine, ensuring that job creation and monitoring are handled concurrently but within the limits of `maxConcurrentJobs`. This prevents the system from being overwhelmed by too many simultaneous Kubernetes API calls.

### Graceful Shutdown

The `Runner` also includes a `quit` channel. This channel is used to signal the `Runner` to stop processing new tasks and gracefully shut down, ensuring that all active jobs are completed before the application exits.

## Usage

Here is a basic example of how to use the `k8sJobs` package:

```go
package main

import (
    "context"
    "github.com/zcubbs/spark/pkg/k8s/jobs"
)

func main() {
    ctx := context.Background()
    kubeconfig := "/path/to/kubeconfig"

    // Initialize the Runner with a limit of 5 concurrent jobs
    runner, err := k8sJobs.New(ctx, kubeconfig, 5)
    if err != nil {
        panic(err)
    }

    // Add tasks to the queue
    runner.AddTask(k8sJobs.Task{ID: "task1", Command: []string{"echo", "Hello World"}, Name: "echo-task", Image: "busybox"})
    runner.AddTask(k8sJobs.Task{ID: "task2", Command: []string{"sleep", "10"}, Name: "sleep-task", Image: "busybox"})

    // Implementing a signal handling or a similar mechanism is advised to call runner.Shutdown() when the application is stopping.
}
```

## Conclusion

The k8sJobs package simplifies the management of Kubernetes jobs, leveraging Go’s concurrency features to provide a robust, easy-to-use interface for running tasks within a Kubernetes cluster. It ensures efficient and safe execution of jobs, making it an ideal tool for applications requiring automated task management in Kubernetes environments.

