package redis

import (
	"crypto/tls"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const envconfigPrefix = "REDIS"

// config represents common configuration options for a Redis connection
type config struct {
	Host      string `envconfig:"HOST" required:"true"`
	Port      int    `envconfig:"PORT" default:"6379"`
	Password  string `envconfig:"PASSWORD"`
	DB        int    `envconfig:"DB"`
	EnableTLS bool   `envconfig:"ENABLE_TLS"`
	Prefix    string `envconfig:"PREFIX"`
}

// Client returns a connection to a Redis database specified by environment
// variables
func Client() (*redis.Client, error) {
	c := config{}
	err := envconfig.Process(envconfigPrefix, &c)
	if err != nil {
		return nil, errors.Wrap(
			err,
			"error getting redis configuration from environment",
		)
	}

	redisOpts := &redis.Options{
		Addr:       fmt.Sprintf("%s:%d", c.Host, c.Port),
		Password:   c.Password,
		DB:         c.DB,
		MaxRetries: 5, // TODO: Should this be configurable?
	}
	if c.EnableTLS {
		redisOpts.TLSConfig = &tls.Config{
			ServerName: c.Host,
		}
	}

	return redis.NewClient(redisOpts), nil
}
