package meta

import (
	"time"
)

const APIVersion = "github.com/krancour/brignext/v2"

type TypeMeta struct {
	Kind       string `json:"kind" bson:"kind"`
	APIVersion string `json:"apiVersion" bson:"apiVersion"`
}

// nolint: lll
type ObjectMeta struct {
	ID string `json:"id,omitempty" bson:"id"`
	// The JSON schema doesn't permit the fields below to be set via the API.
	Created *time.Time `json:"created,omitempty" bson:"created"`
	// CreatedBy   string     `json:"createdBy,omitempty" bson:"createdBy"`
	LastUpdated *time.Time `json:"lastUpdated,omitempty" bson:"lastUpdated"`
	// LastUpdatedBy string     `json:"lastUpdatedBy,omitempty" bson:"lastUpdatedBy"`
}

// nolint: lll
type ObjectReferenceMeta struct {
	ID      string    `json:"id,omitempty" bson:"id"`
	Created time.Time `json:"created,omitempty" bson:"created"`
}
