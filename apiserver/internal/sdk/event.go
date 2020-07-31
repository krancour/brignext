package sdk

import (
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
	"go.mongodb.org/mongo-driver/bson"
)

type Event struct {
	meta.TypeMeta   `json:",inline" bson:",inline"`
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
	Worker     *WorkerSpec       `json:"worker,omitempty" bson:"worker"`
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
	Canceled   *time.Time        `json:"canceled,omitempty" bson:"canceled"`
	Status     *EventStatus      `json:"status,omitempty" bson:"status"`
}

// TODO: Add EventSpec type

func NewEvent() Event {
	return Event{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "Event",
		},
	}
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
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`
	Items         []Event `json:"items"`
}

type EventReference struct {
	meta.TypeMeta            `json:",inline"`
	meta.ObjectReferenceMeta `json:"metadata" bson:",inline"`
	ProjectID                string           `json:"projectID" bson:"projectID"`
	Source                   string           `json:"source" bson:"source"`
	Type                     string           `json:"type" bson:"type"`
	Kubernetes               KubernetesConfig `json:"-" bson:"kubernetes"`
	WorkerPhase              WorkerPhase      `json:"workerPhase" bson:"-"`
}

func NewEventReference(event Event) EventReference {
	eventRef := EventReference{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "EventReference",
		},
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
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`
	Items         []EventReference `json:"items"`
}

func NewEventReferenceList() EventReferenceList {
	return EventReferenceList{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "EventReferenceList",
		},
		ListMeta: meta.ListMeta{},
		Items:    []EventReference{},
	}
}
