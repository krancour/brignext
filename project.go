package brignext

// nolint: lll
type Project struct {
	TypeMeta   `json:",inline" bson:",inline"`
	ObjectMeta `json:"metadata" bson:"metadata"`
	Spec       ProjectSpec            `json:"spec" bson:"spec"`
	Kubernetes *ProjectKubernetesMeta `json:"kubernetes,omitempty" bson:"kubernetes"`
}

// nolint: lll
type ProjectSpec struct {
	Description        string              `json:"description" bson:"description"`
	EventSubscriptions []EventSubscription `json:"eventSubscriptions" bson:"eventSubscriptions"`
	WorkerConfig       WorkerConfig        `json:"workerConfig" bson:"workerConfig"`
}

type ProjectKubernetesMeta struct {
	Namespace string `json:"namespace" bson:"namespace"`
}
