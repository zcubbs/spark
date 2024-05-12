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
	t.Status = "PENDING"

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
			go func(t Task) {
				defer r.wg.Done()
				r.handleTask(ctx, t)
			}(task)
		case <-r.quit:
			return
		}
	}
}

// handleTask handles a task by creating a job, monitoring it, and updating the task status.
func (r *Runner) handleTask(parentCtx context.Context, t Task) {
	// Acquire a semaphore slot before processing a task
	r.currentJobs <- struct{}{}        // This will block if the limit is reached
	defer func() { <-r.currentJobs }() // Release the semaphore slot after the task is processed

	if t.Timeout == 0 {
		t.Timeout = r.defaultJobTimeout
	}

	// Create a new context with a timeout for the task
	taskCtx, cancel := context.WithTimeout(parentCtx, time.Duration(t.Timeout)*time.Second+5*time.Second)
	defer cancel()

	t.StartedAt = time.Now()
	log.Debug("Processing task", "jobId", t.ID, "image", t.Image, "command", t.Command)
	t.Status = "RUNNING"
	r.updateTaskStatus(t)

	if _, err := r.createAndMonitorJob(taskCtx, r.namespace, t); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Warn("Task timed out", "jobId", t.ID)
			t.Status = "TIMEOUT"
			t.Logs = fmt.Sprintf("Task timed out after %d seconds", t.Timeout)
		} else {
			log.Error("Failed to create or monitor job", "error", err, "jobId", t.ID)
			t.Status = "FAILED"
			t.Logs = fmt.Sprintf("Job monitoring failed: %v", err)
		}
	} else {
		t.Status = "SUCCEEDED"
		t.Logs, err = r.getLogs(taskCtx, t.ID) // Retrieve logs
		if err != nil {
			log.Error("Failed to get logs", "error", err, "jobId", t.ID)
		}
	}

	t.EndedAt = time.Now()
	r.updateTaskStatus(t)
	r.delete(taskCtx, t.ID) // Use task-specific context for deletion
}

func (r *Runner) updateTaskStatus(t Task) bool {
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
