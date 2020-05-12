package brignext

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// nolint: lll
type Event struct {
	ID         string                 `json:"id,omitempty" bson:"_id"`
	ProjectID  string                 `json:"projectID" bson:"projectID"`
	Source     string                 `json:"source" bson:"source"`
	Type       string                 `json:"type" bson:"type"`
	Labels     EventLabels            `json:"labels" bson:"labels"`
	ShortTitle string                 `json:"shortTitle" bson:"shortTitle"`
	LongTitle  string                 `json:"longTitle" bson:"longTitle"`
	Git        EventGitConfig         `json:"git" bson:"git"`
	Kubernetes *EventKubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
	Worker     *Worker                `json:"worker,omitempty" bson:"worker"`
	Payload    string                 `json:"payload,omitempty" bson:"-"`
	Created    *time.Time             `json:"created,omitempty" bson:"created"`
	// TODO: These fields are not yet in use
	CreatedBy  string `json:"createdBy,omitempty" bson:"createdBy"`
	CanceledBy string `json:"canceledBy,omitempty" bson:"canceledBy"`
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
