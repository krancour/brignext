package brignext

type ProjectList struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata"`
	Items    []Project `json:"items"`
}

func NewProjectList() ProjectList {
	return ProjectList{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "ProjectList",
		},
		Items: []Project{},
	}
}

type Project struct {
	TypeMeta   `json:",inline" bson:",inline"`
	ObjectMeta `json:"metadata" bson:"metadata"`
	Spec       ProjectSpec `json:"spec" bson:"spec"`
	// The JSON schema doesn't permit the fields below to be set via the API.
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
}

// nolint: lll
type ProjectSpec struct {
	Description        string              `json:"description" bson:"description"`
	EventSubscriptions []EventSubscription `json:"eventSubscriptions" bson:"eventSubscriptions"`
	Worker             WorkerSpec          `json:"worker" bson:"worker"`
}
