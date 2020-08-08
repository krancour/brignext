package meta

import "time"

// APIVersion represents the API and major version thereof with which this
// version of the BrigNext API server is compatible.
const APIVersion = "github.com/krancour/brignext/v2"

// TypeMeta represents metadata about a resource type to help clients and
// servers mutually head off potential confusion over types (and versions of
// thereof) sent over the wire.
type TypeMeta struct {
	// Kind specifies the type of a serialized resource.
	Kind string `json:"kind,omitempty"`
	// APIVersion specifiesthe major version of the BrigNext API with which the
	// client or server having serialized the resource is compatible.
	APIVersion string `json:"apiVersion,omitempty"`
}

// ObjectMeta represents metadata about an instance of a resource. The fields
// of this type are broadly applicable to most if not all resource types.
type ObjectMeta struct {
	// ID is an immutable resource identifier.
	ID string `json:"id,omitempty" bson:"id,omitempty"`
	// Created indicates the time at which a resource was created.
	Created *time.Time `json:"created,omitempty" bson:"created,omitempty"`
	// CreatedBy indicates who or what created the resource.
	CreatedBy *PrincipalReference `json:"createdBy,omitempty" bson:"createdBy,omitempty"` // nolint: lll
	// LastUpdated indicates the time at which a resource was last updated.
	LastUpdated *time.Time `json:"lastUpdated,omitempty" bson:"lastUpdated,omitempty"` // nolint: lll
	// LastUpdatedBy indicates who or what last updated the resource.
	LastUpdatedBy *PrincipalReference `json:"lastUpdatedBy,omitempty" bson:"lastUpdatedBy,omitempty"` // nolint: lll
}

// PrincipalReference is a reference to a user of the system-- either a human
// user or service account.
type PrincipalReference struct {
	// User references a human User by their ID.
	User string `json:"user,omitempty" bson:"user,omitempty"`
	// ServiceAccount references a ServiceAccount by its ID.
	ServiceAccount string `json:"serviceAcount,omitempty" bson:"serviceAcount,omitempty"` // nolint: lll
}

// ObjectReferenceMeta is an abridged representation of resource metadata used
// by other abridged representations of various resource types.
type ObjectReferenceMeta struct {
	ID      string    `json:"id,omitempty" bson:"id"`
	Created time.Time `json:"created,omitempty" bson:"created"`
}
