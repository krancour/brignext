package sdk

import (
	"encoding/json"

	"github.com/krancour/brignext/v2/sdk/meta"
)

// Project is Brignext's fundamental management construct. Through a
// ProjectSpec, it pairs EventSubscriptions with a template WorkerSpec.
type Project struct {
	// ObjectMeta encapsulates Project metadata.
	meta.ObjectMeta `json:"metadata"`
	// Description is a natural language description of the Project.
	Description string `json:"description"`
	// Spec is an instance of a ProjectSpec that pairs EventSubscriptions with
	// a WorkerTemplate.
	Spec ProjectSpec `json:"spec"`
	// Kubernetes encapsulates Kubernetes-specific details of the Project's
	// environment. These details are populated by BrigNext so that sufficiently
	// authorized Kubernetes users may obtain the information needed to directly
	// modify a Project's environment to facilitate certain advanced use cases.
	// Clients MUST leave the value of this field nil when using the API to create
	// or update a Project.
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty"`
}

// MarshalJSON amends Project instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
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
	EventSubscriptions []EventSubscription `json:"eventSubscriptions"`
	// WorkerTemplate is a prototypical WorkerSpec.
	WorkerTemplate WorkerSpec `json:"workerTemplate"`
}

// EventSubscription defines a set of Events of interest. ProjectSpecs utilize
// these in defining the events that should trigger the execution of a new
// Worker. An Event matches a subscription if it meets ALL of the specified
// criteria.
type EventSubscription struct {
	// Source specifies the origin of an Event (e.g. a gateway).
	Source string `json:"source"`
	// Types enumerates specific Events of interest from the specified source.
	// This is useful in narrowing a subscription when a source also emits many
	// events that are NOT of interest.
	Types []string `json:"types"`
	// Labels enumerates specific key/value pairs with which Events of interest
	// must be labeled. An event must have ALL of these labels to match this
	// subscription.
	Labels Labels `json:"labels"`
}

// KubernetesConfig represents Kubernetes-specific configuration. This is used
// primarily at the Project level, but is also denormalized onto Events so that
// Event handling doesn't required a Project lookup to obtain
// Kubernetes-specific configuration.
type KubernetesConfig struct {
	Namespace string `json:"namespace"`
}

// Secret represents Project-level sensitive information.
type Secret struct {
	// Key is a key by which the secret can referred.
	Key string `json:"key"`
	// Value is the sensitive information.
	Value string `json:"value"`
}

// MarshalJSON amends Secret instances with type metadata so that clients do not
// need to be concerned with the tedium of doing so.
func (s Secret) MarshalJSON() ([]byte, error) {
	type Alias Secret
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "Secret",
			},
			Alias: (Alias)(s),
		},
	)
}
