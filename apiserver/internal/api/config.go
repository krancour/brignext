package api

import (
	"errors"

	"github.com/kelseyhightower/envconfig"
	"github.com/krancour/brignext/v2/apiserver/internal/crypto"
)

const envconfigPrefix = "API_SERVER"

// We use an exported interface to govern access to our config because the
// underlying struct has fields we don't want to expose.
type Config interface {
	Port() int
	RootUserEnabled() bool
	HashedRootUserPassword() string
	HashedSchedulerToken() string
	HashedObserverToken() string
	TLSEnabled() bool
	TLSCertPath() string
	TLSKeyPath() string
}

type config struct {
	PortAttr                   int    `envconfig:"PORT"`
	RootUserEnabledAttr        bool   `envconfig:"ROOT_USER_ENABLED"`
	RootUserPasswordAttr       string `envconfig:"ROOT_USER_PASSWORD"`
	HashedRootUserPasswordAttr string
	SchedulerTokenAttr         string `envconfig:"SCHEDULER_TOKEN" required:"true"` // nolint: lll
	HashedSchedulerTokenAttr   string
	ObserverTokenAttr          string `envconfig:"OBSERVER_TOKEN" required:"true"` // nolint: lll
	HashedObserverTokenAttr    string
	TLSEnabledAttr             bool   `envconfig:"TLS_ENABLED"`
	TLSCertPathAttr            string `envconfig:"TLS_CERT_PATH"`
	TLSKeyPathAttr             string `envconfig:"TLS_KEY_PATH"`
}

// NewConfigWithDefaults returns a Config object with default values already
// applied. Callers are then free to set custom values for the remaining fields
// and/or override default values.
func NewConfigWithDefaults() Config {
	return &config{PortAttr: 8080}
}

// GetConfigFromEnvironment returns configuration derived from environment
// variables
func GetConfigFromEnvironment() (Config, error) {
	c := NewConfigWithDefaults().(*config)
	if err := envconfig.Process(envconfigPrefix, c); err != nil {
		return c, err
	}

	if c.RootUserEnabledAttr && c.RootUserPasswordAttr == "" {
		return c, errors.New(
			"with the root user enabled, a value is required for the " +
				"ROOT_USER_PASSWORD environment variable",
		)
	}

	if c.TLSEnabledAttr {
		if c.TLSCertPathAttr == "" {
			return c, errors.New(
				"with TLS enabled, a value is required for the " +
					"TLS_CERT_PATH environment variable",
			)
		}
		if c.TLSKeyPathAttr == "" {
			return c, errors.New(
				"with TLS enabled, a value is required for the " +
					"TLS_KEY_PATH environment variable",
			)
		}
	}

	c.HashedRootUserPasswordAttr = crypto.ShortSHA("root", c.RootUserPasswordAttr)
	// Don't let the unencrypted password float around in memory!
	c.RootUserPasswordAttr = ""

	c.HashedSchedulerTokenAttr = crypto.ShortSHA("", c.SchedulerTokenAttr)
	// Don't let the unencrypted token float around in memory!
	c.SchedulerTokenAttr = ""

	c.HashedObserverTokenAttr = crypto.ShortSHA("", c.ObserverTokenAttr)
	// Don't let the unencrypted token float around in memory!
	c.ObserverTokenAttr = ""

	return c, nil
}

func (c *config) Port() int {
	return c.PortAttr
}

func (c *config) RootUserEnabled() bool {
	return c.RootUserEnabledAttr
}

func (c *config) HashedRootUserPassword() string {
	return c.HashedRootUserPasswordAttr
}

func (c *config) HashedSchedulerToken() string {
	return c.HashedSchedulerTokenAttr
}

func (c *config) HashedObserverToken() string {
	return c.HashedObserverTokenAttr
}

func (c *config) TLSEnabled() bool {
	return c.TLSEnabledAttr
}

func (c *config) TLSCertPath() string {
	return c.TLSCertPathAttr
}

func (c *config) TLSKeyPath() string {
	return c.TLSKeyPathAttr
}
