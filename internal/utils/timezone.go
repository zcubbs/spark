package utils

import (
	"github.com/charmbracelet/log"
	"time"
)

// CheckTimeZone checks the current timezone and time
func CheckTimeZone() {
	log.Info("server time",
		"tz", time.Now().Location().String(),
		"now", time.Now().Format("2006-01-02T15:04:05.000 MST"))
}
