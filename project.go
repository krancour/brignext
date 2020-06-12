package brignext

import (
	"github.com/krancour/brignext/v2/internal/pkg/meta"
	"go.mongodb.org/mongo-driver/bson"
)

type Project struct {
	meta.TypeMeta   `json:",inline" bson:",inline"`
	meta.ObjectMeta `json:"metadata" bson:"metadata"`
	Spec            ProjectSpec `json:"spec" bson:"spec"`
	// The JSON schema doesn't permit the fields below to be set via the API.
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
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
	Description        string              `json:"description" bson:"description"`
	EventSubscriptions []EventSubscription `json:"eventSubscriptions" bson:"eventSubscriptions"` // nolint: lll
	// TODO: Consider renaming this field to WorkerTemplate
	Worker WorkerSpec `json:"worker" bson:"worker"`
}

type EventSubscription struct {
	Source string   `json:"source" bson:"source"`
	Types  []string `json:"types" bson:"types"`
	Labels Labels   `json:"labels" bson:"labels"`
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
