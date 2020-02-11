package brignext

import "time"

type WorkerStatus string

const (
	WorkerStatusPending WorkerStatus = "PENDING"
)

// nolint: lll
type Worker struct {
	ID            string       `json:"id,omitempty" bson:"_id,omitempty"`
	ProjectID     string       `json:"projectID,omitempty" bson:"projectID,omitempty"`
	EventID       string       `json:"eventID,omitempty" bson:"eventID,omitempty"`
	EventProvider string       `json:"eventProvider,omitempty" bson:"eventProvider,omitempty"`
	EventType     string       `json:"eventType,omitempty" bson:"eventType,omitempty"`
	Image         Image        `json:"image,omitempty" bson:"image,omitempty"`
	Command       string       `json:"command,omitempty" bson:"command,omitempty"`
	Status        WorkerStatus `json:"status,omitempty" bson:"status,omitempty"`
	Created       time.Time    `json:"created,omitempty" bson:"created,omitempty"`
	Started       *time.Time   `json:"started,omitempty" bson:"started,omitempty"`
	Ended         *time.Time   `json:"ended,omitempty" bson:"ended,omitempty"`
	ExitCode      int32        `json:"exitCode,omitempty" bson:"exitCode,omitempty"`
}
