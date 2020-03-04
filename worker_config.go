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
	TriggeringEvents []TriggeringEvents     `json:"events" bson:"events"`
	Container        ContainerConfig        `json:"container" bson:"container"`
	WorkspaceSize    string                 `json:"workspaceSize" bson:"workspaceSize"`
	Git              WorkerGitConfig        `json:"git" bson:"git"`
	Kubernetes       WorkerKubernetesConfig `json:"kubernetes" bson:"kubernetes"`
	Jobs             JobsConfig             `json:"jobs" bson:"jobs"`
	LogLevel         LogLevel               `json:"logLevel" bson:"logLevel"`
}

type TriggeringEvents struct {
	Provider string   `json:"provider" bson:"provider"`
	Types    []string `json:"types" bson:"types"`
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
