package brignext

import "time"

type ObjectMeta struct {
	ID      string     `json:"id" bson:"id"`
	Created *time.Time `json:"created,omitempty" bson:"created"`
	// TODO: These fields are not yet in use
	CreatedBy     string     `json:"createdBy,omitempty" bson:"createdBy"`
	LastUpdated   *time.Time `json:"lastUpdated,omitempty" bson:"lastUpdated"`
	LastUpdatedBy string     `json:"lastUpdatedBy,omitempty" bson:"lastUpdatedBy"`
}
