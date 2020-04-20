package brignext

import (
	"time"
)

type LogEntry struct {
	Time    time.Time `json:"time" bson:"time"`
	Message string    `json:"message" bson:"log"`
}
