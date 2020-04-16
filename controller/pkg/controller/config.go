package controller

import "github.com/kelseyhightower/envconfig"

const envconfigPrefix = "CONTROLLER"

type Config struct {
	APIAddress            string `envconfig:"API_ADDRESS" required:"true"`
	APIToken              string `envconfig:"API_TOKEN" required:"true"`
	IgnoreAPICertWarnings bool   `envconfig:"IGNORE_API_CERT_WARNINGS"`
	WorkspaceStorageClass string `envconfig:"WORKSPACE_STORAGE_CLASS"`
}

// NewConfigWithDefaults returns a Config object with default values already
// applied. Callers are then free to set custom values for the remaining fields
// and/or override default values.
func NewConfigWithDefaults() Config {
	return Config{}
}

// GetConfigFromEnvironment returns configuration derived from environment
// variables
func GetConfigFromEnvironment() (Config, error) {
	c := NewConfigWithDefaults()
	err := envconfig.Process(envconfigPrefix, &c)
	return c, err
}
