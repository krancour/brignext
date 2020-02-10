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
	Password  string `envconfig:"PASSWORD" required:"true"`
	DB        int    `envconfig:"DB" required:"true"`
	EnableTLS bool   `envconfig:"ENABLE_TLS" default:"false"`
}

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
		MaxRetries: 5,
	}
	if c.EnableTLS {
		redisOpts.TLSConfig = &tls.Config{
			ServerName: c.Host,
		}
	}

	return redis.NewClient(redisOpts), nil
}
