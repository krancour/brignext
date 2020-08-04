package sdk

import (
	"encoding/json"

	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
	"go.mongodb.org/mongo-driver/bson"
)

// Project is Brignext's fundamental management construct. Through a
// ProjectSpec, it pairs EventSubscriptions with a template WorkerSpec.
type Project struct {
	// ObjectMeta contains Project metadata.
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	// Description is a natural language description of the Project.
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	// Spec is an instance of a ProjectSpec that pairs EventSubscriptions with
	// a WorkerTemplate.
	Spec ProjectSpec `json:"spec" bson:"spec"`
	// Kubernetes contains Kubernetes-specific details of the Project's
	// environment.
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes,omitempty"` // nolint: lll
}

// MarshalJSON amends Project instances with type metadata.
func (p Project) MarshalJSON() ([]byte, error) {
	type Alias Project
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "Project",
			},
			Alias: (Alias)(p),
		},
	)
}

// ProjectSpec is the technical component of a Project. It pairs
// EventSubscriptions with a prototypical WorkerSpec that is used as a template
// for creating new Workers.
type ProjectSpec struct {
	// EventSubscription defines a set of trigger conditions under which a new
	// Worker should be created.
	EventSubscriptions []EventSubscription `json:"eventSubscriptions,omitempty" bson:"eventSubscriptions,omitempty"` // nolint: lll
	// WorkerTemplate is a prototypical WorkerSpec.
	WorkerTemplate WorkerSpec `json:"workerTemplate" bson:"workerTemplate"`
}

// EventSubscription defines a set of Events of interest. ProjectSpecs utilize
// these in defining the Events that should trigger the execution of a new
// Worker. An Event matches a subscription if it meets ALL of the specified
// criteria.
type EventSubscription struct {
	// Source specifies the origin of an Event (e.g. a gateway).
	Source string `json:"source,omitempty" bson:"source,omitempty"`
	// Types enumerates specific Events of interest from the specified source.
	// This is useful in narrowing a subscription when a source also emits many
	// events that are NOT of interest.
	Types []string `json:"types,omitempty" bson:"types,omitempty"`
	// Labels enumerates specific key/value pairs with which Events of interest
	// must be labeled. An event must have ALL of these labels to match this
	// subscription.
	Labels Labels `json:"labels,omitempty" bson:"labels,omitempty"`
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

// KubernetesConfig represents Kubernetes-specific configuration. This is used
// primarily at the Project level, but is also denormalized onto Events so that
// Event handling doesn't required a Project lookup to obtain
// Kubernetes-specific configuration.
type KubernetesConfig struct {
	// Namespace is the dedicated Kubernetes namespace for the Project. This is
	// NOT specified by clients when creating a new Project. The namespace is
	// created by / assigned by the system. This detail is a necessity to prevent
	// clients from naming existing namespaces in an attempt to hijack them.
	Namespace string `json:"namespace,omitempty" bson:"namespace,omitempty"`
}

// ProjectReference is an abridged representation of a Project useful to
// API operations that construct and return potentially large collections of
// projects. Utilizing such an abridged representation both limits response size
// and accounts for the reality that not all clients with authorization to list
// projects are authorized to view the details of every Project.
type ProjectReference struct {
	// ObjectReferenceMeta contains an abridged representation of Project
	// metadata.
	meta.ObjectReferenceMeta `json:"metadata" bson:",inline"`
	// Description is a natural language description of the Project.
	Description string `json:"description,omitempty" bson:"description,omitempty"`
}

// MarshalJSON amends ProjectReference instances with type metadata.
func (p ProjectReference) MarshalJSON() ([]byte, error) {
	type Alias ProjectReference
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ProjectReference",
			},
			Alias: (Alias)(p),
		},
	)
}

// ProjectReferenceList is an ordered list of ProjectReferences.
type ProjectReferenceList struct {
	// Items is a slice of ProjectReferences.
	//
	// TODO: When pagination is implemented, list metadata will need to be added
	Items []ProjectReference `json:"items"`
}

// MarshalJSON amends ProjectReferenceList instances with type metadata.
func (p ProjectReferenceList) MarshalJSON() ([]byte, error) {
	type Alias ProjectReferenceList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ProjectReferenceList",
			},
			Alias: (Alias)(p),
		},
	)
}
