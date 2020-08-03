package meta

import (
	"time"
)

const APIVersion = "github.com/krancour/brignext/v2"

type TypeMeta struct {
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

// nolint: lll
type ObjectMeta struct {
	ID string `json:"id,omitempty" bson:"id,omitempty"`
	// The JSON schema doesn't permit the fields below to be set via the API.
	Created *time.Time `json:"created,omitempty" bson:"created,omitempty"`
	// CreatedBy   string     `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	LastUpdated *time.Time `json:"lastUpdated,omitempty" bson:"lastUpdated,omitempty"`
	// LastUpdatedBy string     `json:"lastUpdatedBy,omitempty" bson:"lastUpdatedBy,omitempty"`
}

// nolint: lll
type ObjectReferenceMeta struct {
	ID      string    `json:"id,omitempty" bson:"id"`
	Created time.Time `json:"created,omitempty" bson:"created"`
}
