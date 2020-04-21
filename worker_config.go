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
	TriggeringEvents     []TriggeringEvents     `json:"events" bson:"events"`
	Container            ContainerConfig        `json:"container" bson:"container"`
	WorkspaceSize        string                 `json:"workspaceSize" bson:"workspaceSize"`
	Git                  WorkerGitConfig        `json:"git" bson:"git"`
	Kubernetes           WorkerKubernetesConfig `json:"kubernetes" bson:"kubernetes"`
	JobsConfig           JobsConfig             `json:"jobsConfig" bson:"jobsConfig"`
	LogLevel             LogLevel               `json:"logLevel" bson:"logLevel"`
	ConfigFilesDirectory string                 `json:"configFilesDirectory" bson:"configFilesDirectory"`
	// TODO: Add support for embedding files right here in the worker config
}

type TriggeringEvents struct {
	Source string   `json:"source" bson:"source"`
	Types  []string `json:"types" bson:"types"`
}

func (w *WorkerConfig) Matches(eventSource, eventType string) bool {
	if len(w.TriggeringEvents) == 0 {
		return true
	}
	for _, tes := range w.TriggeringEvents {
		if tes.Matches(eventSource, eventType) {
			return true
		}
	}
	return false
}

func (t *TriggeringEvents) Matches(eventSource, eventType string) bool {
	if t.Source == "" ||
		eventSource == "" ||
		eventType == "" ||
		t.Source != eventSource {
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
