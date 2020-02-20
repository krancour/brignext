package brignext

import (
	"time"
)

type LogEntry struct {
	Time    time.Time
	Message string
}
