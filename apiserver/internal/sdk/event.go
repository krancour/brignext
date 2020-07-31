package sdk

import (
	"encoding/json"

	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
	"go.mongodb.org/mongo-driver/bson"
)

type Event struct {
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	ProjectID       string         `json:"projectID" bson:"projectID"`
	Source          string         `json:"source" bson:"source"`
	Type            string         `json:"type" bson:"type"`
	Labels          Labels         `json:"labels" bson:"labels"`
	ShortTitle      string         `json:"shortTitle" bson:"shortTitle"`
	LongTitle       string         `json:"longTitle" bson:"longTitle"`
	Git             EventGitConfig `json:"git" bson:"git"`
	Payload         string         `json:"payload" bson:"payload"`
	// The JSON schema doesn't permit the fields below to be set via the API.
	// TODO: Some of these don't need to be pointers anymore
	Worker     *WorkerSpec       `json:"worker,omitempty" bson:"worker"`
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
	Status     *EventStatus      `json:"status,omitempty" bson:"status"`
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

// UnmarshalBSON implements custom BSON unmarshaling for the Event type.
// This does little more than guarantees that the Labels field isn't nil so that
// custom unmarshaling of the Labels (which is more involved) can succeed.
func (e *Event) UnmarshalBSON(bytes []byte) error {
	if e.Labels == nil {
		e.Labels = Labels{}
	}
	type EventAlias Event
	return bson.Unmarshal(
		bytes,
		&struct {
			*EventAlias `bson:",inline"`
		}{
			EventAlias: (*EventAlias)(e),
		},
	)
}

type EventGitConfig struct {
	CloneURL string `json:"cloneURL" bson:"cloneURL"`
	Commit   string `json:"commit" bson:"commit"`
	Ref      string `json:"ref" bson:"ref"`
}

type EventStatus struct {
	WorkerStatus WorkerStatus         `json:"workerStatus" bson:"workerStatus"`
	JobStatuses  map[string]JobStatus `json:"jobStatuses" bson:"jobStatuses"`
}

type EventListOptions struct {
	ProjectID    string
	WorkerPhases []WorkerPhase
}

type EventList struct {
	Items []Event `json:"items"`
}

type EventReference struct {
	meta.ObjectReferenceMeta `json:"metadata" bson:",inline"`
	ProjectID                string           `json:"projectID" bson:"projectID"`
	Source                   string           `json:"source" bson:"source"`
	Type                     string           `json:"type" bson:"type"`
	Kubernetes               KubernetesConfig `json:"-" bson:"kubernetes"`
	WorkerPhase              WorkerPhase      `json:"workerPhase" bson:"-"`
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

func NewEventReference(event Event) EventReference {
	eventRef := EventReference{
		ObjectReferenceMeta: meta.ObjectReferenceMeta{
			ID:      event.ID,
			Created: *event.Created,
		},
		ProjectID: event.ProjectID,
	}
	eventRef.Source = event.Source
	eventRef.Type = event.Type
	if event.Kubernetes != nil {
		eventRef.Kubernetes = *event.Kubernetes
	}
	if event.Status != nil {
		eventRef.WorkerPhase = event.Status.WorkerStatus.Phase
	}
	return eventRef
}

func (e *EventReference) UnmarshalBSON(bytes []byte) error {
	event := Event{}
	if err := bson.Unmarshal(bytes, &event); err != nil {
		return err
	}
	*e = NewEventReference(event)
	return nil
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
