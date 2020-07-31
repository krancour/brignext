package sdk

import (
	"encoding/json"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type Project struct {
	meta.ObjectMeta `json:"metadata"`
	Description     string      `json:"description"`
	Spec            ProjectSpec `json:"spec"`
	// The JSON schema doesn't permit the fields below to be set via the API, so
	// they are pointers. Their values must be nil when outbound.
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty"`
}

// TODO: Add ProjectStatus type-- move KubernetesConfig under there? Maybe?

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
	meta.ObjectReferenceMeta `json:"metadata"`
	Description              string `json:"description"`
}

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

type ProjectReferenceList struct {
	Items []ProjectReference `json:"items"`
}

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
