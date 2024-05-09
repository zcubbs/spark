package k8sJobs

import (
	"github.com/tidwall/buntdb"
	"strings"
)

func (r *Runner) GetLogsForTaskFromDB(taskId string) (string, error) {
	var logs string
	err := r.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(taskId)
		if err != nil {
			return err
		}
		logs = strings.Split(val, ",")[4]
		return nil
	})
	if err != nil {
		return "", err
	}
	return logs, nil
}
