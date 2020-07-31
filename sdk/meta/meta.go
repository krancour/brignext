package meta

import (
	"time"
)

const APIVersion = "github.com/krancour/brignext/v2"

type TypeMeta struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
}

// nolint: lll
type ObjectMeta struct {
	ID string `json:"id,omitempty"`
	// The JSON schema doesn't permit the fields below to be set via the API.
	// Pointers must be nil and strings must be empty when outbound.
	Created *time.Time `json:"created,omitempty"`
	// CreatedBy   string     `json:"createdBy,omitempty"`
	LastUpdated *time.Time `json:"lastUpdated,omitempty"`
	// LastUpdatedBy string     `json:"lastUpdatedBy,omitempty"`
}

// nolint: lll
type ObjectReferenceMeta struct {
	ID      string    `json:"id,omitempty"`
	Created time.Time `json:"created,omitempty"`
}
