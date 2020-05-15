package brignext

import (
	"go.mongodb.org/mongo-driver/bson"
)

type EventSubscription struct {
	Source string   `json:"source" bson:"source"`
	Types  []string `json:"types" bson:"types"`
	Labels Labels   `json:"labels" bson:"labels"`
}

// UnmarshalBSON implements custom BSON unmarshaling for the EventSubscription
// type. This does little more than guarantees that the Labels field isn't nil
// so that custom unmarshaling of the EventLabels (which is more involved) can
// succeed.
func (e *EventSubscription) UnmarshalBSON(bytes []byte) error {
	if e.Labels == nil {
		e.Labels = Labels{}
	}
	type EventSubscriptionAlias EventSubscription
	return bson.Unmarshal(
		bytes,
		&struct {
			*EventSubscriptionAlias `bson:",inline"`
		}{
			EventSubscriptionAlias: (*EventSubscriptionAlias)(e),
		},
	)
}
