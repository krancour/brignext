package sdk

import (
	"github.com/krancour/brignext/v2/sdk/meta"
)

type Project struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata"`
	Description     string      `json:"description"`
	Spec            ProjectSpec `json:"spec"`
	// The JSON schema doesn't permit the fields below to be set via the API, so
	// they are pointers. Their values must be nil when outbound.
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty"`
}

// TODO: Add ProjectStatus type-- move KubernetesConfig under there

func NewProject() Project {
	return Project{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "Project",
		},
	}
}

type ProjectSpec struct {
	EventSubscriptions []EventSubscription `json:"eventSubscriptions"`
	// TODO: Consider renaming this field to WorkerTemplate
	Worker WorkerSpec `json:"worker"`
}

type EventSubscription struct {
	Source string   `json:"source"`
	Types  []string `json:"types"`
	Labels Labels   `json:"labels"`
}

type ProjectReference struct {
	meta.TypeMeta            `json:",inline"`
	meta.ObjectReferenceMeta `json:"metadata"`
	Description              string `json:"description"`
}

type ProjectReferenceList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`
	Items         []ProjectReference `json:"items"`
}
