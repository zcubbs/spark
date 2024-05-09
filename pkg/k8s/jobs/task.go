package k8sJobs

import (
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/tidwall/buntdb"
	"strings"
	"time"
)

// Task represents a task to be executed by the runner.
type Task struct {
	ID      string
	Command []string
	Image   string
	Timeout int
}

// AddTask adds a task to the runner.
func (r *Runner) AddTask(t Task) error {
	err := r.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(t.ID, fmt.Sprintf("%s,%s,%s,QUEUED", t.ID, t.Image, strings.Join(t.Command, " ")), nil)
		return err
	})
	if err != nil {
		return err
	}

	select {
	case r.taskChan <- t:
		return nil // Successfully added the task
	default:
		// Handle the case when taskChan is full
		return fmt.Errorf("task queue is full")
	}
}

// processTasks processes tasks from the task channel.
func (r *Runner) processTasks(ctx context.Context) {
	for {
		select {
		case task := <-r.taskChan:
			r.wg.Add(1)
			go r.handleTask(ctx, task)
		case <-r.quit:
			return
		}
	}
}

func (r *Runner) handleTask(ctx context.Context, t Task) {
	defer r.wg.Done()

	// Set initial task status to RUNNING
	if !(r.updateTaskStatus(t.ID, t.Image, t.Command, "RUNNING", "")) {
		// If the task status update fails, return
		return
	}

	// Process the task with a timeout
	jobCtx, cancel := context.WithTimeout(ctx, time.Duration(t.Timeout)*time.Second)
	defer cancel()

	_, err := r.createAndMonitorJob(jobCtx, r.namespace, t)
	if err != nil {
		if errors.Is(jobCtx.Err(), context.DeadlineExceeded) {
			r.handleTimeout(t)
		} else {
			log.Error("Failed to create and monitor job", "error", err, "jobId", t.ID)
			r.updateTaskStatus(t.ID, t.Image, t.Command, "FAILED", fmt.Sprintf("Job failed: %v", err))
		}
		r.delete(ctx, t.ID) // Cleanup after failure or timeout
		return
	}

	// Retrieve logs and update task status to SUCCEEDED or FAILED based on log retrieval status
	logs, err := r.GetLogs(ctx, t.ID)
	finalStatus := "SUCCEEDED"
	if err != nil {
		finalStatus = "FAILED"
		logs = fmt.Sprintf("Failed to get logs: %v", err)
	}

	r.updateTaskStatus(t.ID, t.Image, t.Command, finalStatus, logs)
	r.delete(ctx, t.ID) // Cleanup after successful completion
}

func (r *Runner) updateTaskStatus(jobId, image string, command []string, status, logs string) bool {
	err := r.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(jobId, fmt.Sprintf("%s,%s,%s,%s,%s",
			jobId, image, strings.Join(command, " "), status, logs), nil)
		return err
	})
	if err != nil {
		log.Error("Failed to update DB task status", "error", err, "jobId", jobId)
	}

	return err == nil
}

func (r *Runner) handleTimeout(t Task) {
	r.updateTaskStatus(t.ID, t.Image, t.Command, "TIMED OUT", "Job timed out")
	r.delete(context.Background(), t.ID) // Ensure context.Background() to avoid passing a canceled context
}
