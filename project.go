package brignext

import (
	"time"
)

// nolint: lll
type Project struct {
	ID                 string                   `json:"id" bson:"_id"`
	Description        string                   `json:"description" bson:"description"`
	EventSubscriptions []EventSubscription      `json:"eventSubscriptions" bson:"eventSubscriptions"`
	WorkerConfig       WorkerConfig             `json:"workerConfig" bson:"workerConfig"`
	Kubernetes         *ProjectKubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
	Created            *time.Time               `json:"created,omitempty" bson:"created"`
	// TODO: These fields are not yet in use
	CreatedBy     string `json:"createdBy,omitempty" bson:"createdBy"`
	LastUpdatedBy string `json:"lastUpdatedBy,omitempty" bson:"lastUpdatedBy"`
}
