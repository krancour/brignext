package brignext

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type EventList struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata"`
	Items    []Event `json:"items"`
}

type Event struct {
	TypeMeta   `json:",inline" bson:",inline"`
	ObjectMeta `json:"metadata" bson:"metadata"`
	ProjectID  string         `json:"projectID" bson:"projectID"`
	Source     string         `json:"source" bson:"source"`
	Type       string         `json:"type" bson:"type"`
	Labels     Labels         `json:"labels" bson:"labels"`
	ShortTitle string         `json:"shortTitle" bson:"shortTitle"`
	LongTitle  string         `json:"longTitle" bson:"longTitle"`
	Git        EventGitConfig `json:"git" bson:"git"`
	Payload    string         `json:"payload" bson:"-"`
	// The JSON schema doesn't permit the fields below to be set via the API.
	Worker     *WorkerSpec       `json:"worker,omitempty" bson:"worker"`
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
	Canceled   *time.Time        `json:"canceled,omitempty" bson:"canceled"`
	Status     *EventStatus      `json:"status,omitempty" bson:"status"`
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
