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
	Container            ContainerConfig        `json:"container" bson:"container"`
	WorkspaceSize        string                 `json:"workspaceSize" bson:"workspaceSize"`
	Git                  WorkerGitConfig        `json:"git" bson:"git"`
	Kubernetes           WorkerKubernetesConfig `json:"kubernetes" bson:"kubernetes"`
	JobsConfig           JobsConfig             `json:"jobsConfig" bson:"jobsConfig"`
	LogLevel             LogLevel               `json:"logLevel" bson:"logLevel"`
	ConfigFilesDirectory string                 `json:"configFilesDirectory" bson:"configFilesDirectory"`
	DefaultConfigFiles   map[string]string      `json:"defaultConfigFiles" bson:"defaultConfigFiles"`
}
