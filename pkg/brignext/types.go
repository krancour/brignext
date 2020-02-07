package brignext

import (
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	FirstSeen time.Time `json:"firstSeen"`
}

type ServiceAccount struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
}

type Session struct {
	ID            string    `json:"id"`
	Root          bool      `json:"root"`
	Authenticated bool      `json:"authenticated"`
	UserID        string    `json:"userID"`
	ExpiresAt     time.Time `json:"expiresAt"`
}

type Project struct {
	Name                 string       `json:"name"`
	Repo                 Repo         `json:"repo"`
	DefaultScript        string       `json:"defaultScript"`
	DefaultScriptName    string       `json:"defaultScriptName"`
	DefaultConfig        string       `json:"defaultConfig"`
	DefaultConfigName    string       `json:"defaultConfigName"`
	Kubernetes           Kubernetes   `json:"kubernetes"`
	SharedSecret         string       `json:"sharedSecret"`
	Github               Github       `json:"github"`
	Secrets              SecretsMap   `json:"secrets"`
	Worker               WorkerConfig `json:"worker"`
	InitGitSubmodules    bool         `json:"initGitSubmodules"`
	AllowPrivilegedJobs  bool         `json:"allowPrivilegedJobs"`
	AllowHostMounts      bool         `json:"allowHostMounts"`
	ImagePullSecrets     string       `json:"imagePullSecrets"`
	WorkerCommand        string       `json:"workerCommand"`
	BrigadejsPath        string       `json:"brigadejsPath"`
	BrigadeConfigPath    string       `json:"brigadeConfigPath"`
	GenericGatewaySecret string       `json:"genericGatewaySecret"`
}

type Repo struct {
	Name     string `json:"name"`
	CloneURL string `json:"cloneURL"`
	SSHKey   string `json:"sshKey"`
	SSHCert  string `json:"sshCert"`
}

type Kubernetes struct {
	Namespace         string `json:"namespace"`
	VCSSidecar        string `json:"vcsSidecar"`
	BuildStorageSize  string `json:"buildStorageSize"`
	BuildStorageClass string `json:"buildStorageClass"`
	CacheStorageClass string `json:"cacheStorageClass"`
	AllowSecretKeyRef bool   `json:"allowSecretKeyRef"`
	ServiceAccount    string `json:"serviceAccount"`
}

type Github struct {
	Token     string `json:"token"`
	BaseURL   string `json:"baseURL"`
	UploadURL string `json:"uploadURL"`
}

type SecretsMap map[string]interface{}

type WorkerConfig struct {
	Registry   string `json:"registry"`
	Name       string `json:"name"`
	Tag        string `json:"tag"`
	PullPolicy string `json:"pullPolicy"`
}

type Build struct {
	ID          string    `json:"id"`
	ProjectName string    `json:"projectName"`
	Type        string    `json:"type"`
	Provider    string    `json:"provider"`
	ShortTitle  string    `json:"shortTitle"`
	LongTitle   string    `json:"longTitle"`
	CloneURL    string    `json:"cloneURL"`
	Revision    *Revision `json:"revision"`
	Payload     []byte    `json:"payload"`
	Script      []byte    `json:"script"`
	Config      []byte    `json:"config"`
	Worker      *Worker   `json:"worker"`
	LogLevel    string    `json:"logLevel"`
}

type Revision struct {
	Commit string `json:"commit"`
	Ref    string `json:"ref"`
}

type Worker struct {
	ID        string    `json:"id"`
	BuildID   string    `json:"buildID"`
	ProjectID string    `json:"projectID"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	ExitCode  int32     `json:"exitCode"`
	Status    JobStatus `json:"status"`
}

type JobStatus string

const (
	JobPending   JobStatus = "Pending"
	JobRunning   JobStatus = "Running"
	JobSucceeded JobStatus = "Succeeded"
	JobFailed    JobStatus = "Failed"
	JobUnknown   JobStatus = "Unknown"
)

type LogEntry struct {
	Time    time.Time
	Message string
}
