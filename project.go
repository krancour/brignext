package brignext

import (
	"time"
)

// nolint: lll
type Project struct {
	ID               string                   `json:"id" bson:"_id"`
	Description      string                   `json:"description" bson:"description"`
	TriggeringEvents []TriggeringEvents       `json:"events" bson:"events"`
	WorkerConfig     WorkerConfig             `json:"workerConfig" bson:"workerConfig"`
	Kubernetes       *ProjectKubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
	Created          *time.Time               `json:"created,omitempty" bson:"created"`
}

func (p *Project) Matches(eventSource, eventType string) bool {
	if len(p.TriggeringEvents) == 0 {
		return true
	}
	for _, tes := range p.TriggeringEvents {
		if tes.Matches(eventSource, eventType) {
			return true
		}
	}
	return false
}
