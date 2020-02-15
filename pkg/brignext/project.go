package brignext

import (
	"time"
)

// nolint: lll
type Project struct {
	ID          string                  `json:"id" bson:"_id,omitempty"`
	Description string                  `json:"description,omitempty" bson:"description,omitempty"`
	Workers     map[string]WorkerConfig `json:"workers,omitempty" bson:"workers,omitempty"`
	Namespace   string                  `json:"namespace,omitempty" bson:"namespace,omitempty"`
	Created     *time.Time              `json:"created,omitempty" bson:"created,omitempty"`
	// Repo        Repo   `json:"repo,omitempty" bson:"repo,omitempty"`
	// DefaultScript     string     `json:"defaultScript,omitempty" bson:"defaultScript,omitempty"`
	// DefaultScriptName string     `json:"defaultScriptName,omitempty" bson:"defaultScriptName,omitempty"`
	// DefaultConfig     string     `json:"defaultConfig,omitempty" bson:"defaultConfig,omitempty"`
	// DefaultConfigName string     `json:"defaultConfigName,omitempty" bson:"defaultConfigName,omitempty"`
	// Github       Github `json:"github,omitempty" bson:"github,omitempty"`
	// // TODO: We MUST encrypt these!
	// Secrets              map[string]string `json:"secrets,omitempty" bson:"secrets,omitempty"`
	// InitGitSubmodules    bool              `json:"initGitSubmodules,omitempty" bson:"initGitSubmodules,omitempty"`
	// ImagePullSecrets     string            `json:"imagePullSecrets,omitempty" bson:"imagePullSecrets,omitempty"`
	// BrigadejsPath        string            `json:"brigadejsPath,omitempty" bson:"brigadejsPath,omitempty"`
	// BrigadeConfigPath    string            `json:"brigadeConfigPath,omitempty" bson:"brigadeConfigPath,omitempty"`
}

func (p *Project) GetWorkers(
	eventProvider string,
	eventType string,
) map[string]Worker {
	workers := map[string]Worker{}
	for workerName, workerConfig := range p.Workers {
		if workerConfig.Matches(eventProvider, eventType) {
			workers[workerName] = Worker{
				InitContainer: workerConfig.InitContainer,
				Container:     workerConfig.Container,
				Jobs:          workerConfig.Jobs,
				Kubernetes:    workerConfig.Kubernetes,
				Status:        WorkerStatusPending,
			}
		}
	}
	return workers
}

// type Repo struct {
// 	// TODO: The name field may actually be useless here
// 	// Name     string `json:"name,omitempty" bson:"name,omitempty"`
// 	CloneURL string `json:"cloneURL" bson:"cloneURL"`
// 	// // TODO: We MUST encrypt this!
// 	// SSHKey  string `json:"sshKey,omitempty" bson:"sshKey,omitempty"`
// 	// SSHCert string `json:"sshCert,omitempty" bson:"sshCert,omitempty"`
// }

// type Github struct {
// 	// TODO: We MUST encrypt this!
// 	Token     string `json:"token,omitempty" bson:"token,omitempty"`
// 	BaseURL   string `json:"baseURL,omitempty" bson:"baseURL,omitempty"`
// 	UploadURL string `json:"uploadURL,omitempty" bson:"uploadURL,omitempty"`
// }
