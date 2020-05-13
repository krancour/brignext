package brignext

type TypeMeta struct {
	Kind       string `json:"kind" bson:"kind"`
	APIVersion string `json:"apiVersion" bson:"apiVersion"`
}
