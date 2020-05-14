package brignext

import "time"

// nolint: lll
type Project struct {
	TypeMeta    `json:",inline" bson:",inline"`
	ProjectMeta `json:"metadata" bson:"metadata"`
	Spec        ProjectSpec            `json:"spec" bson:"spec"`
	Kubernetes  *ProjectKubernetesMeta `json:"kubernetes,omitempty" bson:"kubernetes"`
}

type ProjectMeta struct {
	ID      string     `json:"id" bson:"id"`
	Created *time.Time `json:"created,omitempty" bson:"created"`
	// TODO: These fields are not yet in use
	CreatedBy     string     `json:"createdBy,omitempty" bson:"createdBy"`
	LastUpdated   *time.Time `json:"lastUpdated,omitempty" bson:"lastUpdated"`
	LastUpdatedBy string     `json:"lastUpdatedBy,omitempty" bson:"lastUpdatedBy"`
}

// nolint: lll
type ProjectSpec struct {
	Description        string              `json:"description" bson:"description"`
	EventSubscriptions []EventSubscription `json:"eventSubscriptions" bson:"eventSubscriptions"`
	Worker             WorkerSpec          `json:"worker" bson:"worker"`
}

type ProjectKubernetesMeta struct {
	Namespace string `json:"namespace" bson:"namespace"`
}
