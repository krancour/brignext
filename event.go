package brignext

import "time"

type EventStatus string

const (
	// EventStatusMoot represents the state wherein an event has been accepted by
	// the system, but (per project configuration) triggered no constituent
	// workers.
	EventStatusMoot EventStatus = "MOOT"
	// EventStatusPending represents the state wherein an event has been accepted
	// by the system, but all of (n > 1) workers remain in a PENDING state.
	EventStatusPending EventStatus = "PENDING"
	// EventStatusCanceled represents the state wherein a PENDING event was
	// canceled.
	EventStatusCanceled EventStatus = "CANCELED"
	// EventStatusProcessing represents the state wherein an event is currently
	// being processed (i.e. at least one constituent worker has entered a RUNNING
	// or terminal state AND NOT ALL constituent workers have entered a terminal
	// state).
	EventStatusProcessing EventStatus = "PROCESSING"
	// EventStatusAborted represents the state wherein a PROCESSING event was
	// forcefully terminated.
	EventStatusAborted EventStatus = "ABORTED"
	// EventStatusSucceeded represents the state wherein all constituent workers
	// have entered a SUCCEEDED state.
	EventStatusSucceeded EventStatus = "SUCCEEDED"
	// EventStatusSucceeded represents the state wherein all constituent workers
	// have entered a terminal state and at least on constituent worker has
	// entered a FAILED state.
	EventStatusFailed EventStatus = "FAILED"
)

type Event struct {
	ID         string                 `json:"id,omitempty" bson:"_id"`
	ProjectID  string                 `json:"projectID" bson:"projectID"`
	Provider   string                 `json:"provider" bson:"provider"`
	Type       string                 `json:"type" bson:"type"`
	ShortTitle string                 `json:"shortTitle" bson:"shortTitle"`
	LongTitle  string                 `json:"longTitle" bson:"longTitle"`
	Git        EventGitConfig         `json:"git" bson:"git"`
	Status     EventStatus            `json:"status,omitempty" bson:"status"`
	Kubernetes *EventKubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
	Workers    map[string]Worker      `json:"workers,omitempty" bson:"workers"`
	Created    *time.Time             `json:"created,omitempty" bson:"created"`
	// TODO: This should be encrypted! Maybe?
	// Payload  []byte    `json:"payload,omitempty" bson:"payload"`
	// Script   []byte    `json:"script,omitempty" bson:"script"`
	// Config   []byte    `json:"config,omitempty" bson:"config,"`
}
