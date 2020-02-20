package brignext

import "time"

type EventStatus string

const (
	// EventStatusAccepted represents the state wherein an event has been accepted
	// by the system, but all (n > 1) constituent workers remain pending
	// execution.
	EventStatusAccepted EventStatus = "ACCEPTED"
	// EventStatusMoot represents the state wherein an event has been accepted by
	// the system, but (per project configuration) triggered no constituent
	// workers.
	EventStatusMoot EventStatus = "MOOT"
	// EventStatusCanceled represents the state wherein an accepted event was
	// canceled prior to any processing being performed (i.e. prior to any
	// constituent workers entering a running state).
	EventStatusCanceled EventStatus = "CANCELED"
	// EventStatusProcessing represents the state wherein an event is currently
	// being processed (i.e. at least one constituent worker has entered a running
	// state AND not all constituent workers have entered a terminal state).
	EventStatusProcessing EventStatus = "PROCESSING"
	// EventStatusAborted represents the state wherein event processing was
	// forcefully terminated.
	EventStatusAborted EventStatus = "ABORTED"
	// EventStatusSucceeded represents the state wherein all constituent workers
	// have entered a success state.
	EventStatusSucceeded EventStatus = "SUCCEEDED"
	// EventStatusSucceeded represents the state wherein at least one constituent
	// workers has entered a failed state.
	EventStatusFailed EventStatus = "FAILED"
)

// nolint: lll
type Event struct {
	ID         string            `json:"id,omitempty" bson:"_id,omitempty"`
	ProjectID  string            `json:"projectID,omitempty" bson:"projectID,omitempty"`
	Provider   string            `json:"provider,omitempty" bson:"provider,omitempty"`
	Type       string            `json:"type,omitempty" bson:"type,omitempty"`
	ShortTitle string            `json:"shortTitle,omitempty" bson:"shortTitle,omitempty"`
	LongTitle  string            `json:"longTitle,omitempty" bson:"longTitle,omitempty"`
	Status     EventStatus       `json:"status,omitempty" bson:"status,omitempty"`
	Namespace  string            `json:"namespace,omitempty" bson:"namespace,omitempty"`
	Workers    map[string]Worker `json:"workers,omitempty" bson:"workers,omitempty"`
	Created    *time.Time        `json:"created,omitempty" bson:"created,omitempty"`
	// CloneURL string    `json:"cloneURL,omitempty" bson:"cloneURL,omitempty"`
	// Revision *Revision `json:"revision,omitempty" bson:"revision,omitempty"`
	// TODO: This should be encrypted!
	// Payload  []byte    `json:"payload,omitempty" bson:"payload,omitempty"`
	// Script   []byte    `json:"script,omitempty" bson:"script,omitempty"`
	// Config   []byte    `json:"config,omitempty" bson:"config,omitempty"`
	// LogLevel string    `json:"logLevel,omitempty" bson:"logLevel,omitempty"`
}

// type Revision struct {
// 	Commit string `json:"commit,omitempty" bson:"commit,omitempty"`
// 	Ref    string `json:"ref,omitempty" bson:"ref,omitempty"`
// }
