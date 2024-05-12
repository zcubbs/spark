package k8sJobs

import (
	"encoding/json"
	"github.com/charmbracelet/log"
	"github.com/tidwall/buntdb"
)

func (r *Runner) GetLogsForTaskFromDB(taskId string) (string, error) {
	var logs string
	err := r.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(taskId)
		if err != nil {
			return err
		}
		var task Task
		if err := json.Unmarshal([]byte(val), &task); err != nil {
			return err
		}
		logs = task.Logs
		return nil
	})
	if err != nil {
		return "", err
	}
	return logs, nil
}

func (r *Runner) GetStatusForTaskFromDB(taskId string) (string, error) {
	var status string
	err := r.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(taskId)
		if err != nil {
			return err
		}
		var task Task
		if err := json.Unmarshal([]byte(val), &task); err != nil {
			return err
		}
		status = task.Status
		return nil
	})
	if err != nil {
		return "", err
	}
	return status, nil
}

func (r *Runner) GetTaskFromDB(taskId string) (*Task, error) {
	var task Task
	err := r.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(taskId)
		if err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(val), &task); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *Runner) GetAllTasksFromDB() ([]Task, error) {
	var tasks []Task
	err := r.db.View(func(tx *buntdb.Tx) error {
		return tx.Descend("by_started", func(key, value string) bool {
			var task Task
			if err := json.Unmarshal([]byte(value), &task); err != nil {
				log.Error("Failed to deserialize task", "error", err)
				return true // continue processing
			}
			tasks = append(tasks, task)
			return len(tasks) < 50 // stop after 50
		})
	})
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
