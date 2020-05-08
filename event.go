package brignext

import "time"

// nolint: lll
type Event struct {
	ID         string                 `json:"id,omitempty" bson:"_id"`
	ProjectID  string                 `json:"projectID" bson:"projectID"`
	Source     string                 `json:"source" bson:"source"`
	Type       string                 `json:"type" bson:"type"`
	ShortTitle string                 `json:"shortTitle" bson:"shortTitle"`
	LongTitle  string                 `json:"longTitle" bson:"longTitle"`
	Git        EventGitConfig         `json:"git" bson:"git"`
	Kubernetes *EventKubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
	Worker     *Worker                `json:"worker,omitempty" bson:"worker"`
	Created    *time.Time             `json:"created,omitempty" bson:"created"`
	Payload    string                 `json:"payload,omitempty" bson:"-"`
}
