package brignext

import "github.com/krancour/brignext/v2/internal/pkg/meta"

type ProjectList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`
	Items         []Project `json:"items"`
}

func NewProjectList() ProjectList {
	return ProjectList{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "ProjectList",
		},
		Items: []Project{},
	}
}

type Project struct {
	meta.TypeMeta   `json:",inline" bson:",inline"`
	meta.ObjectMeta `json:"metadata" bson:"metadata"`
	Spec            ProjectSpec `json:"spec" bson:"spec"`
	// The JSON schema doesn't permit the fields below to be set via the API.
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
}

// nolint: lll
type ProjectSpec struct {
	Description        string              `json:"description" bson:"description"`
	EventSubscriptions []EventSubscription `json:"eventSubscriptions" bson:"eventSubscriptions"`
	// TODO: Consider renaming this field to WorkerTemplate
	Worker WorkerSpec `json:"worker" bson:"worker"`
}
