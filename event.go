package brignext

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// nolint: lll
type Event struct {
	TypeMeta   `json:",inline" bson:",inline"`
	EventMeta  `json:"metadata" bson:"metadata"`
	Spec       EventSpec            `json:"spec" bson:"spec"`
	Kubernetes *EventKubernetesMeta `json:"kubernetes,omitempty" bson:"kubernetes"`
	Status     *EventStatus         `json:"status,omitempty" bson:"status"`
}

type EventMeta struct {
	ID         string      `json:"id,omitempty" bson:"id"`
	ProjectID  string      `json:"projectID,omitempty" bson:"projectID"`
	Source     string      `json:"source" bson:"source"`
	Type       string      `json:"type" bson:"type"`
	ShortTitle string      `json:"shortTitle" bson:"shortTitle"`
	LongTitle  string      `json:"longTitle" bson:"longTitle"`
	Labels     EventLabels `json:"labels" bson:"labels"`
	Created    *time.Time  `json:"created,omitempty" bson:"created"`
	// TODO: These fields are not yet in use
	CreatedBy  string     `json:"createdBy,omitempty" bson:"createdBy"`
	Canceled   *time.Time `json:"canceled,omitempty" bson:"canceled"`
	CanceledBy string     `json:"canceledBy,omitempty" bson:"canceledBy"`
}

type EventSpec struct {
	Git     EventGitConfig `json:"git" bson:"git"`
	Worker  *WorkerSpec    `json:"worker,omitempty" bson:"worker"`
	Payload string         `json:"payload,omitempty" bson:"-"`
}

type EventKubernetesMeta struct {
	Namespace string `json:"namespace" bson:"namespace"`
}

type EventStatus struct {
	WorkerStatus WorkerStatus         `json:"workerStatus" bson:"workerStatus"`
	JobStatuses  map[string]JobStatus `json:"jobStatuses" bson:"jobStatuses"`
}

// UnmarshalBSON implements custom BSON marshaling for the Event type. This does
// little more than guarantees that the Labels field isn't nil so that custom
// marshaling of the EventLabels (which is more involved) can succeed.
func (e *Event) UnmarshalBSON(bytes []byte) error {
	if e.Labels == nil {
		e.Labels = EventLabels{}
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
