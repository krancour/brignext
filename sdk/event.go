package sdk

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type Event struct {
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

func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "Event",
			},
			Alias: (Alias)(e),
		},
	)
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
	meta.ObjectReferenceMeta `json:"metadata"`
	ProjectID                string      `json:"projectID"`
	Source                   string      `json:"source"`
	Type                     string      `json:"type"`
	WorkerPhase              WorkerPhase `json:"workerPhase"`
}

func (e EventReference) MarshalJSON() ([]byte, error) {
	type Alias EventReference
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "EventReference",
			},
			Alias: (Alias)(e),
		},
	)
}

type EventReferenceList struct {
	Items []EventReference `json:"items"`
}

func (e EventReferenceList) MarshalJSON() ([]byte, error) {
	type Alias EventReferenceList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "EventReferenceList",
			},
			Alias: (Alias)(e),
		},
	)
}
