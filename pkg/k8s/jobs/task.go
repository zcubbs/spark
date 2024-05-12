package k8sJobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/tidwall/buntdb"
	"time"
)

// Task represents a task to be executed by the runner.
type Task struct {
	ID        string    `json:"id"`
	Command   []string  `json:"command"`
	Image     string    `json:"image"`
	Timeout   int       `json:"timeout"`
	CreatedAt time.Time `json:"created_at"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
	Status    string    `json:"status"`
	Logs      string    `json:"logs"`
}

// AddTask adds a task to the runner.
func (r *Runner) AddTask(t Task) error {
	t.CreatedAt = time.Now() // Set the creation time when the task is added

	taskData, err := json.Marshal(t)
	if err != nil {
		return err
	}

	err = r.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(t.ID, string(taskData), nil)
		return err
	})
	if err != nil {
		return err
	}

	select {
	case r.taskChan <- t:
		return nil // Successfully added the task
	default:
		return fmt.Errorf("task queue is full")
	}
}

// processTasks processes tasks from the task channel.
func (r *Runner) processTasks(ctx context.Context) {
	for {
		select {
		case task := <-r.taskChan:
			r.wg.Add(1)
			// Acquire a semaphore slot before processing a task
			r.currentJobs <- struct{}{} // This will block if the limit is reached
			go func(t Task) {
				defer r.wg.Done()
				defer func() { <-r.currentJobs }() // Release the semaphore slot after the task is processed
				r.handleTask(ctx, t)
			}(task)
		case <-r.quit:
			return
		}
	}
}

func (r *Runner) handleTask(ctx context.Context, t Task) {
	defer r.wg.Done()

	// set default timeout
	if t.Timeout == 0 {
		t.Timeout = r.defaultJobTimeout
	}

	t.StartedAt = time.Now() // Record when the task processing starts
	log.Debug("Processing task", "jobId", t.ID, "image", t.Image, "command", t.Command)

	// Update task status to RUNNING
	if !r.updateTaskStatus(t, "RUNNING", "") {
		return
	}

	// Attempt to create and monitor the Kubernetes job
	_, err := r.createAndMonitorJob(ctx, r.namespace, t)
	if err != nil {
		t.EndedAt = time.Now() // Set end time when task finishes or fails
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			r.handleTimeout(t)
		} else {
			log.Error("Failed to create and monitor job", "error", err, "jobId", t.ID)
			r.updateTaskStatus(t, "FAILED", fmt.Sprintf("Job failed: %v", err))
		}
		r.delete(ctx, t.ID) // Cleanup after failure or timeout
		return
	}

	// Successfully completed the task
	t.EndedAt = time.Now()
	log.Debug("Retrieving logs", "jobId", t.ID)
	logs, err := r.getLogs(ctx, t.ID)
	finalStatus := "SUCCEEDED"
	if err != nil {
		finalStatus = "FAILED"
		logs = fmt.Sprintf("Failed to get logs: %v", err)
		log.Error("Failed to get logs", "error", err, "jobId", t.ID)
	}

	r.updateTaskStatus(t, finalStatus, logs)
	r.delete(ctx, t.ID) // Cleanup after successful completion
}

func (r *Runner) updateTaskStatus(t Task, status, logs string) bool {
	t.Status = status
	t.Logs = logs

	taskData, err := json.Marshal(t)
	if err != nil {
		log.Error("Failed to serialize task", "error", err, "jobId", t.ID)
		return false
	}

	err = r.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(t.ID, string(taskData), nil)
		return err
	})
	if err != nil {
		log.Error("Failed to update DB task status", "error", err, "jobId", t.ID)
		return false
	}

	return true
}

func (r *Runner) handleTimeout(t Task) {
	t.EndedAt = time.Now() // Set the timeout end time
	r.updateTaskStatus(t, "TIMED OUT", "Task timed out")
	r.delete(context.Background(), t.ID) // Ensure context.Background() to avoid passing a canceled context
}
