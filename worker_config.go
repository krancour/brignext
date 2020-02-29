package brignext

type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// nolint: lll
type WorkerConfig struct {
	TriggeringEvents []TriggeringEvents `json:"events,omitempty" bson:"events,omitempty"`
	Container        *ContainerConfig   `json:"container,omitempty" bson:"container,omitempty"`
	WorkspaceSize    string             `json:"workspaceSize,omitempty" bson:"workspaceSize,omitempty"`
	Git              *GitConfig         `json:"git,omitempty" bson:"git,omitempty"`
	Kubernetes       *KubernetesConfig  `json:"kubernetes,omitempty" bson:"kubernetes,omitempty"`
	Jobs             *JobsConfig        `json:"jobs,omitempty" bson:"jobs,omitempty"`
	LogLevel         LogLevel           `json:"logLevel,omitempty" bson:"logLevel,omitempty"`
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
