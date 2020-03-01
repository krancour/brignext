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
	TriggeringEvents []TriggeringEvents      `json:"events,omitempty" bson:"events"`
	Container        *ContainerConfig        `json:"container,omitempty" bson:"container"`
	WorkspaceSize    string                  `json:"workspaceSize,omitempty" bson:"workspaceSize"`
	Git              *GitConfig              `json:"git,omitempty" bson:"git"`
	Kubernetes       *WorkerKubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
	Jobs             *JobsConfig             `json:"jobs,omitempty" bson:"jobs"`
	LogLevel         LogLevel                `json:"logLevel,omitempty" bson:"logLevel"`
}

type TriggeringEvents struct {
	Provider string   `json:"provider,omitempty" bson:"provider"`
	Types    []string `json:"types,omitempty" bson:"types"`
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
