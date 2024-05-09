# `k8sJobs` Go Package

The `k8sJobs` package is a Go library for managing Kubernetes jobs. It provides functionalities for creating, monitoring, and deleting jobs, managing concurrency, and maintaining a record of job statuses and logs in a local BuntDB database.

## Features

- **Dynamic Job Management**: Create and monitor Kubernetes jobs dynamically within your application.
- **Concurrency Control**: Manage multiple jobs concurrently with a configurable limit.
- **Task Queuing**: Queue tasks with a channel-based mechanism.
- **Local Persistence**: Utilize BuntDB to store job statuses and logs.
- **Timeout Handling**: Automatically handle job execution with configurable timeouts.
- **Error Handling**: Robust error handling throughout the job lifecycle.

## Structure

### Runner Struct

Orchestrates job tasks and interacts with the Kubernetes API.

- **Fields**:
    - `cs`: Kubernetes ClientSet to interact with Kubernetes API.
    - `maxConcurrentJobs`: Maximum number of jobs that can run concurrently.
    - `taskChan`: Channel for queuing tasks.
    - `quit`: Channel to signal the shutdown of task processing.
    - `namespace`: Namespace in Kubernetes where jobs are deployed.
    - `db`: BuntDB instance for local data storage.

### Task Struct

Defines the structure of a job task.

- **Fields**:
    - `ID`: Unique identifier of the task.
    - `Command`: Docker container command.
    - `Image`: Docker image used for the job.
    - `Timeout`: Execution timeout in seconds.

## Core Functions

### Initialization

- **`New`**: Initializes a new Runner with specified configuration settings.

### Task Management

- **`AddTask`**: Enqueues a new task for processing.
- **`processTasks`**: Processes tasks from the queue continuously.
- **`handleTask`**: Handles the execution and monitoring of individual tasks.

### Job Monitoring and Management

- **`createAndMonitorJob`**: Sets up and monitors a Kubernetes job.
- **`waitForJobCompletion`**: Waits and checks for the completion of a job.
- **`GetLogs`**: Fetches logs from Kubernetes pods related to the job.
- **`delete`**: Deletes a job from Kubernetes.

### Utility Functions

- **`updateTaskStatus`**: Updates the status of a job in the database.
- **`handleTimeout`**: Handles operations when a job times out.

## Usage Example

```go
// You can replace "default" with the desired path 
// or pass an empty string for the current namespace if in-cluster.
runner, err := k8sJobs.New(context.Background(), "default", 5, 100)
if err != nil {
    log.Fatalf("Failed to start runner: %v", err)
}

task := k8sJobs.Task{
    ID:      "unique-job-id",
    Command: []string{"echo", "Hello World"},
    Image:   "busybox",
    Timeout: 30,
}

if err := runner.AddTask(task); err != nil {
    log.Errorf("Failed to add task: %v", err)
}

// Perform additional operations...

runner.Shutdown()
```
