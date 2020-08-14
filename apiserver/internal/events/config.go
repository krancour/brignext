package events

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/krancour/brignext/v2/apiserver/internal/sdk"
)

const envconfigPrefix = "API_SERVER"

// nolint: lll
type Config struct {
	DefaultWorkerImage           string              `envconfig:"DEFAULT_WORKER_IMAGE"`
	DefaultWorkerImagePullPolicy sdk.ImagePullPolicy `envconfig:"DEFAULT_WORKER_IMAGE_PULL_POLICY"`
	WorkspaceStorageClass        string              `envconfig:"WORKSPACE_STORAGE_CLASS"`
}

func NewConfigWithDefaults() Config {
	return Config{}
}

func GetConfigFromEnvironment() (Config, error) {
	c := NewConfigWithDefaults()
	return c, envconfig.Process(envconfigPrefix, &c)
}
