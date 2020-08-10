package meta

import "time"

// APIVersion represents the API and major version thereof with which this
// version of the BrigNext SDK is compatible.
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
	ID string `json:"id,omitempty"`
	// Created indicates the time at which a resource was created. This is
	// recorded by the system. Clients must leave the value of this field set to
	// nil when using the API to create or update resources.
	Created *time.Time `json:"created,omitempty"`
	// CreatedBy indicates who or what created the resource. This is recorded by
	// the system. Clients must leave the value of this field set to nil when
	// using the API to create or update resources.
	CreatedBy *PrincipalReference `json:"createdBy,omitempty"`
	// LastUpdated indicates the time at which a resource was last updated. This
	// is recorded by the system. Clients must leave the value of this field set
	// to nil when using the API to create or update resources.
	LastUpdated *time.Time `json:"lastUpdated,omitempty"`
	// LastUpdatedBy indicates who or what last updated the resource. This is
	// recorded by the system. Clients must leave the value of this field set to
	// nil when using the API to create or update resources.
	LastUpdatedBy *PrincipalReference `json:"lastUpdatedBy,omitempty"`
}

// PrincipalReference is a reference to a user of the system-- either a human
// user or service account.
type PrincipalReference struct {
	// User references a human User by their ID.
	User string `json:"user,omitempty"`
	// ServiceAccount references a ServiceAccount by its ID.
	ServiceAccount string `json:"serviceAcount,omitempty"`
}

// ListMeta is metadata for ordered collections of resources.
type ListMeta struct {
	// Continue, when non-empty is an opaque value created by and understood by an
	// API operation that returns partial (pageable) results. Submitting this
	// value with subsequent requests to the same operation specifies to that
	// operation which page to return next.
	Continue string `json:"continue,omitempty"`
	// RemainingItemCount, when non-nil indicates that an API operation returned
	// partial (pageable) results and indicates how many results remain.
	RemainingItemCount *int64 `json:"remainingItemCount,omitempty"`
}
