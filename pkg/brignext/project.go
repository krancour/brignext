package brignext

// nolint: lll
type Project struct {
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	Repo        Repo   `json:"repo" bson:"repo"`
	// DefaultScript     string     `json:"defaultScript" bson:"defaultScript"`
	// DefaultScriptName string     `json:"defaultScriptName" bson:"defaultScriptName"`
	// DefaultConfig     string     `json:"defaultConfig" bson:"defaultConfig"`
	// DefaultConfigName string     `json:"defaultConfigName" bson:"defaultConfigName"`
	// Kubernetes        Kubernetes `json:"kubernetes" bson:"kubernetes"`
	// TODO: We MUST encrypt this!
	// SharedSecret string `json:"sharedSecret" bson:"sharedSecret"`
	// Github       Github `json:"github" bson:"github"`
	// // TODO: We MUST encrypt these!
	// Secrets              map[string]string `json:"secrets" bson:"secrets"`
	// Worker               WorkerConfig      `json:"worker" bson:"worker"`
	// InitGitSubmodules    bool              `json:"initGitSubmodules" bson:"initGitSubmodules"`
	// AllowPrivilegedJobs  bool              `json:"allowPrivilegedJobs" bson:"allowPrivilegedJobs"`
	// AllowHostMounts      bool              `json:"allowHostMounts" bson:"allowHostMounts"`
	// ImagePullSecrets     string            `json:"imagePullSecrets" bson:"imagePullSecrets"`
	// WorkerCommand        string            `json:"workerCommand" bson:"workerCommand"`
	// BrigadejsPath        string            `json:"brigadejsPath" bson:"brigadejsPath"`
	// BrigadeConfigPath    string            `json:"brigadeConfigPath" bson:"brigadeConfigPath"`
	// GenericGatewaySecret string            `json:"genericGatewaySecret" bson:"genericGatewaySecret"`
}

type Repo struct {
	// TODO: The name field may actually be useless here
	// Name     string `json:"name" bson:"name"`
	CloneURL string `json:"cloneURL" bson:"cloneURL"`
	// // TODO: We MUST encrypt this!
	// SSHKey  string `json:"sshKey" bson:"sshKey"`
	// SSHCert string `json:"sshCert" bson:"sshCert"`
}

// type Kubernetes struct {
// 	Namespace         string `json:"namespace" bson:"namespace"`
// 	VCSSidecar        string `json:"vcsSidecar" bson:"vcsSidecar"`
// 	BuildStorageSize  string `json:"buildStorageSize" bson:"buildStorageSize"`
// 	BuildStorageClass string `json:"buildStorageClass" bson:"buildStorageClass"`
// 	CacheStorageClass string `json:"cacheStorageClass" bson:"cacheStorageClass"`
// 	AllowSecretKeyRef bool   `json:"allowSecretKeyRef" bson:"allowSecretKeyRef"`
// 	ServiceAccount    string `json:"serviceAccount" bson:"serviceAccount"`
// }

// type Github struct {
// 	// TODO: We MUST encrypt this!
// 	Token     string `json:"token" bson:"token"`
// 	BaseURL   string `json:"baseURL" bson:"baseURL"`
// 	UploadURL string `json:"uploadURL" bson:"uploadURL"`
// }

// type WorkerConfig struct {
// 	Registry   string `json:"registry" bson:"registry"`
// 	Name       string `json:"name" bson:"name"`
// 	Tag        string `json:"tag" bson:"tag"`
// 	PullPolicy string `json:"pullPolicy" bson:"pullPolicy"`
// }
