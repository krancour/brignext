package sdk

import (
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type Event struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata"`
	ProjectID       string         `json:"projectID"`
	Source          string         `json:"source"`
	Type            string         `json:"type"`
	Labels          Labels         `json:"labels"`
	ShortTitle      string         `json:"shortTitle"`
	LongTitle       string         `json:"longTitle"`
	Git             EventGitConfig `json:"git"`
	Payload         string         `json:"payload"`
	// The JSON schema doesn't permit the fields below to be set via the API, so
	// they are pointers. Their values must be nil when outbound.
	Worker     *WorkerSpec       `json:"worker,omitempty"`
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty"`
	Canceled   *time.Time        `json:"canceled,omitempty"`
	Status     *EventStatus      `json:"status,omitempty"`
}

func NewEvent() Event {
	return Event{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "Event",
		},
	}
}

type EventGitConfig struct {
	CloneURL string `json:"cloneURL"`
	Commit   string `json:"commit"`
	Ref      string `json:"ref"`
}

type EventStatus struct {
	WorkerStatus WorkerStatus         `json:"workerStatus"`
	JobStatuses  map[string]JobStatus `json:"jobStatuses"`
}

type EventListOptions struct {
	ProjectID    string
	WorkerPhases []WorkerPhase
}

type EventReference struct {
	meta.TypeMeta            `json:",inline"`
	meta.ObjectReferenceMeta `json:"metadata"`
	ProjectID                string      `json:"projectID"`
	Source                   string      `json:"source"`
	Type                     string      `json:"type"`
	WorkerPhase              WorkerPhase `json:"workerPhase"` // nolint: lll
}

type EventReferenceList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`
	Items         []EventReference `json:"items"`
}
