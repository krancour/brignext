package brignext

import "time"

type EventStatus string

const (
	EventStatusCreated EventStatus = "CREATED"
)

type Event struct {
	ID        string      `json:"id,omitempty" bson:"_id,omitempty"`
	ProjectID string      `json:"projectID,omitempty" bson:"projectID,omitempty"`
	Provider  string      `json:"provider,omitempty" bson:"provider,omitempty"`
	Type      string      `json:"type,omitempty" bson:"type,omitempty"`
	Created   *time.Time  `json:"created,omitempty" bson:"created,omitempty"`
	Status    EventStatus `json:"status,omitempty" bson:"status,omitempty"`
	// ---------------------------------------------------------------------------
	// ShortTitle  string    `json:"shortTitle,omitempty" bson:"shortTitle,omitempty"`
	// LongTitle   string    `json:"longTitle,omitempty" bson:"longTitle,omitempty"`
	// CloneURL string    `json:"cloneURL,omitempty" bson:"cloneURL,omitempty"`
	// Revision *Revision `json:"revision,omitempty" bson:"revision,omitempty"`
	// Payload  []byte    `json:"payload,omitempty" bson:"payload,omitempty"`
	// Script   []byte    `json:"script,omitempty" bson:"script,omitempty"`
	// Config   []byte    `json:"config,omitempty" bson:"config,omitempty"`
	// Worker *Worker `json:"worker,omitempty" bson:"worker,omitempty"`
	// LogLevel string    `json:"logLevel,omitempty" bson:"logLevel,omitempty"`
}

// type Revision struct {
// 	Commit string `json:"commit,omitempty" bson:"commit,omitempty"`
// 	Ref    string `json:"ref,omitempty" bson:"ref,omitempty"`
// }

// type Worker struct {
// 	Status    JobStatus `json:"status,omitempty" bson:"status,omitempty"`
// 	StartTime time.Time `json:"startTime,omitempty" bson:"startTime,omitempty"`
// 	EndTime   time.Time `json:"endTime,omitempty" bson:"endTime,omitempty"`
// 	ExitCode  int32     `json:"exitCode,omitempty" bson:"exitCode,omitempty"`
// }
