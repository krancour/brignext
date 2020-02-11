package brignext

type Image struct {
	Repository string `json:"repository,omitempty" bson:"repository,omitempty"`
	Tag        string `json:"tag,omitempty" bson:"tag,omitempty"`
	PullPolicy string `json:"pullPolicy,omitempty" bson:"pullPolicy,omitempty"`
}
