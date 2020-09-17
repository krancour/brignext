package core

import (
	"github.com/kelseyhightower/envconfig"
)

const envconfigPrefix = "API_SERVER"

type Config struct {
	APIAddress                   string          `envconfig:"API_ADDRESS"`
	DefaultWorkerImage           string          `envconfig:"DEFAULT_WORKER_IMAGE"`             // nolint: lll
	DefaultWorkerImagePullPolicy ImagePullPolicy `envconfig:"DEFAULT_WORKER_IMAGE_PULL_POLICY"` // nolint: lll
	WorkspaceStorageClass        string          `envconfig:"WORKSPACE_STORAGE_CLASS"`          // nolint: lll
}

func NewConfigWithDefaults() Config {
	return Config{}
}

func GetConfigFromEnvironment() (Config, error) {
	c := NewConfigWithDefaults()
	return c, envconfig.Process(envconfigPrefix, &c)
}
