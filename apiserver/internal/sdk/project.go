package sdk

import (
	"encoding/json"

	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
	"go.mongodb.org/mongo-driver/bson"
)

type Project struct {
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	Description     string      `json:"description" bson:"description"`
	Spec            ProjectSpec `json:"spec" bson:"spec"`
	// The JSON schema doesn't permit the fields below to be set via the API.
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
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
	EventSubscriptions []EventSubscription `json:"eventSubscriptions" bson:"eventSubscriptions"` // nolint: lll
	// TODO: Consider renaming this field to WorkerTemplate
	WorkerTemplate WorkerSpec `json:"workerTemplate" bson:"workerTemplate"`
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

type ProjectReference struct {
	meta.ObjectReferenceMeta `json:"metadata" bson:",inline"`
	Description              string `json:"description" bson:"description"`
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
