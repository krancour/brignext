package amqp

import (
	"github.com/brigadecore/brigade/v2/scheduler/internal/queue"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const envconfigPrefix = "AMQP"

// config represents common configuration options for an AMQP connection
type config struct {
	Address           string `envconfig:"ADDRESS" required:"true"`
	Username          string `envconfig:"USERNAME" required:"true"`
	Password          string `envconfig:"PASSWORD" required:"true"`
	IsAzureServiceBus bool   `envconfig:"IS_AZURE_SERVICE_BUS" default:"false"`
}

func GetQueueReaderFactoryFromEnvironment() (queue.ReaderFactory, error) {
	c, err := getConfig()
	if err != nil {
		return nil, err
	}
	return NewQueueReaderFactory(
		c.Address,
		c.Username,
		c.Password,
		c.IsAzureServiceBus,
	)
}

func getConfig() (config, error) {
	c := config{}
	err := envconfig.Process(envconfigPrefix, &c)
	if err != nil {
		return c, errors.Wrap(
			err,
			"error getting AMQP configuration from environment",
		)
	}
	return c, nil
}
