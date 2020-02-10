package brignext

import "time"

// nolint: lll
type Project struct {
	ID            string                  `json:"id" bson:"_id,omitempty"`
	Description   string                  `json:"description,omitempty" bson:"description,omitempty"`
	WorkerConfigs map[string]WorkerConfig `json:"workerConfigs,omitempty" bson:"workerConfigs,omitempty"`
	Created       *time.Time              `json:"created,omitempty" bson:"created,omitempty"`
	// ---------------------------------------------------------------------------
	// Repo        Repo   `json:"repo,omitempty" bson:"repo,omitempty"`
	// DefaultScript     string     `json:"defaultScript,omitempty" bson:"defaultScript,omitempty"`
	// DefaultScriptName string     `json:"defaultScriptName,omitempty" bson:"defaultScriptName,omitempty"`
	// DefaultConfig     string     `json:"defaultConfig,omitempty" bson:"defaultConfig,omitempty"`
	// DefaultConfigName string     `json:"defaultConfigName,omitempty" bson:"defaultConfigName,omitempty"`
	// Kubernetes        Kubernetes `json:"kubernetes,omitempty" bson:"kubernetes,omitempty"`
	// TODO: We MUST encrypt this!
	// SharedSecret string `json:"sharedSecret,omitempty" bson:"sharedSecret,omitempty"`
	// Github       Github `json:"github,omitempty" bson:"github,omitempty"`
	// // TODO: We MUST encrypt these!
	// Secrets              map[string]string `json:"secrets,omitempty" bson:"secrets,omitempty"`
	// InitGitSubmodules    bool              `json:"initGitSubmodules,omitempty" bson:"initGitSubmodules,omitempty"`
	// AllowPrivilegedJobs  bool              `json:"allowPrivilegedJobs,omitempty" bson:"allowPrivilegedJobs,omitempty"`
	// AllowHostMounts      bool              `json:"allowHostMounts,omitempty" bson:"allowHostMounts,omitempty"`
	// ImagePullSecrets     string            `json:"imagePullSecrets,omitempty" bson:"imagePullSecrets,omitempty"`
	// WorkerCommand        string            `json:"workerCommand,omitempty" bson:"workerCommand,omitempty"`
	// BrigadejsPath        string            `json:"brigadejsPath,omitempty" bson:"brigadejsPath,omitempty"`
	// BrigadeConfigPath    string            `json:"brigadeConfigPath,omitempty" bson:"brigadeConfigPath,omitempty"`
	// GenericGatewaySecret string            `json:"genericGatewaySecret,omitempty" bson:"genericGatewaySecret,omitempty"`
}

type WorkerConfig struct {
	Events []string `json:"events,omitempty" bson:"events,omitempty"`
	Image  *Image   `json:"image,omitempty" bson:"image,omitempty"`
}

type Image struct {
	Repository string `json:"repository,omitempty" bson:"repository,omitempty"`
	Tag        string `json:"tag,omitempty" bson:"tag,omitempty"`
	PullPolicy string `json:"pullPolicy,omitempty" bson:"pullPolicy,omitempty"`
}

// type Repo struct {
// 	// TODO: The name field may actually be useless here
// 	// Name     string `json:"name,omitempty" bson:"name,omitempty"`
// 	CloneURL string `json:"cloneURL" bson:"cloneURL"`
// 	// // TODO: We MUST encrypt this!
// 	// SSHKey  string `json:"sshKey,omitempty" bson:"sshKey,omitempty"`
// 	// SSHCert string `json:"sshCert,omitempty" bson:"sshCert,omitempty"`
// }

// type Kubernetes struct {
// 	Namespace         string `json:"namespace,omitempty" bson:"namespace,omitempty"`
// 	VCSSidecar        string `json:"vcsSidecar,omitempty" bson:"vcsSidecar,omitempty"`
// 	BuildStorageSize  string `json:"buildStorageSize,omitempty" bson:"buildStorageSize,omitempty"`
// 	BuildStorageClass string `json:"buildStorageClass,omitempty" bson:"buildStorageClass,omitempty"`
// 	CacheStorageClass string `json:"cacheStorageClass,omitempty" bson:"cacheStorageClass,omitempty"`
// 	AllowSecretKeyRef bool   `json:"allowSecretKeyRef,omitempty" bson:"allowSecretKeyRef,omitempty"`
// 	ServiceAccount    string `json:"serviceAccount,omitempty" bson:"serviceAccount,omitempty"`
// }

// type Github struct {
// 	// TODO: We MUST encrypt this!
// 	Token     string `json:"token,omitempty" bson:"token,omitempty"`
// 	BaseURL   string `json:"baseURL,omitempty" bson:"baseURL,omitempty"`
// 	UploadURL string `json:"uploadURL,omitempty" bson:"uploadURL,omitempty"`
// }
