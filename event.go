package brignext

import "time"

type EventPhase string

const (
	// EventPhasePending represents the state wherein an event has been accepted
	// by the system, but all of (n > 1) workers remain in a PENDING state.
	EventPhasePending EventPhase = "PENDING"
	// EventPhaseCanceled represents the state wherein a PENDING event was
	// canceled.
	EventPhaseCanceled EventPhase = "CANCELED"
	// EventPhaseProcessing represents the state wherein an event is currently
	// being processed (i.e. at least one constituent worker has entered a RUNNING
	// or terminal state AND NOT ALL constituent workers have entered a terminal
	// state).
	EventPhaseProcessing EventPhase = "PROCESSING"
	// EventPhaseAborted represents the state wherein a PROCESSING event was
	// forcefully terminated.
	EventPhaseAborted EventPhase = "ABORTED"
	// EventPhaseSucceeded represents the state wherein all constituent workers
	// have entered a SUCCEEDED state.
	EventPhaseSucceeded EventPhase = "SUCCEEDED"
	// EventPhaseFailed represents the state wherein all constituent workers
	// have entered a terminal state and at least on constituent worker has
	// entered a FAILED state.
	EventPhaseFailed EventPhase = "FAILED"
)

// nolint: lll
type Event struct {
	ID         string                 `json:"id,omitempty" bson:"_id"`
	ProjectID  string                 `json:"projectID" bson:"projectID"`
	Source     string                 `json:"source" bson:"source"`
	Type       string                 `json:"type" bson:"type"`
	ShortTitle string                 `json:"shortTitle" bson:"shortTitle"`
	LongTitle  string                 `json:"longTitle" bson:"longTitle"`
	Git        EventGitConfig         `json:"git" bson:"git"`
	Kubernetes *EventKubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
	Worker     *Worker                `json:"worker,omitempty" bson:"worker"`
	Created    *time.Time             `json:"created,omitempty" bson:"created"`
	Status     *EventStatus           `json:"status,omitempty" bson:"status"`
	Payload    string                 `json:"payload,omitempty" bson:"-"`
}

type EventStatus struct {
	Started *time.Time `json:"started" bson:"started"`
	Ended   *time.Time `json:"ended" bson:"ended"`
	Phase   EventPhase `json:"phase" bson:"phase"`
}
