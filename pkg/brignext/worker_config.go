package brignext

// nolint: lll
type WorkerConfig struct {
	TriggeringEvents []TriggeringEvents `json:"events,omitempty" bson:"events,omitempty"`
	InitContainer    *ContainerConfig   `json:"initContainer,omitempty" bson:"initContainer,omitempty"`
	Container        *ContainerConfig   `json:"container,omitempty" bson:"container,omitempty"`
}

type TriggeringEvents struct {
	Provider string   `json:"provider,omitempty" bson:"provider,omitempty"`
	Types    []string `json:"types,omitempty" bson:"types,omitempty"`
}

func (w *WorkerConfig) Matches(eventProvider, eventType string) bool {
	if len(w.TriggeringEvents) == 0 {
		return true
	}
	for _, tes := range w.TriggeringEvents {
		if tes.Matches(eventProvider, eventType) {
			return true
		}
	}
	return false
}

func (t *TriggeringEvents) Matches(eventProvider, eventType string) bool {
	if t.Provider == "" ||
		eventProvider == "" ||
		eventType == "" ||
		t.Provider != eventProvider {
		return false
	}
	if len(t.Types) == 0 {
		return true
	}
	for _, tipe := range t.Types {
		if tipe == eventType {
			return true
		}
	}
	return false
}
