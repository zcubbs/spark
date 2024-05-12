# `S P A R K`

> Disclaimer: This project is a work in progress and is not yet ready for production use. Please check back later for updates.

`Spark` is a GO application/system designed to manage and monitor Kubernetes jobs efficiently. This project streamlines the process of job execution within a Kubernetes cluster by providing a set of endpoints to schedule, track, and manage jobs programmatically. With integrated local storage, it offers detailed tracking and logging capabilities, making it ideal for environments that require dynamic job scheduling and comprehensive monitoring.

## Features

- **Dynamic Job Scheduling**: Automate the deployment of Kubernetes jobs, leveraging a structured approach to define and manage job specifications such as image, commands, and necessary configurations.
- **Concurrency Management**: Control and limit the number of jobs that can run concurrently, allowing for effective resource utilization and system stability.
- **Task Queuing System**: Utilize an internal queuing system to manage job tasks, ensuring that job submissions are handled efficiently and executed in order.
- **Comprehensive Monitoring**: Continuously monitor the status of each job, capturing and reacting to job completions, failures, and timeouts in real-time.
- **Log Retrieval and Storage**: Automatically fetch and store logs from job executions, providing immediate access to job outputs for debugging and verification purposes.
- **Rate Limiting and Timeouts**: Implement client-side rate limiting and configurable timeouts to manage the load on the Kubernetes API and ensure jobs complete within expected time frames.
- **Local Persistence**: Using BuntDB for fast, in-memory data storage to keep track of job statuses and logs, ensuring data persistence across job operations.

## Usage Scenarios

- **Data processing applications**: Managing batch jobs for data transformation, analysis, or MLM training.
- **General automation**: Running maintenance scripts, backups, and other periodic tasks within a Kubernetes cluster.
- **CI/CD pipelines**: Automating deployment tasks, testing, and other operations that can be encapsulated as Kubernetes jobs.

## Development

### CLI Commands

The CLI provides a set of commands to interact with the `Spark` system. To get started, run the following command:

```bash
go run .\cmd\cli\main.go --mode rest --image "busybox" --cmd "echo Hello World" --timeout 60
```

The command above will create a new job with the specified image, command, and timeout. You can also use the following flags to customize the job:

- `--mode`: The mode of operation (`rest` or `grpc`).
- `--image`: The Docker image to use for the job.
- `--cmd`: The command to run inside the container.
- `--timeout`: The maximum duration for the job to complete.
