package amqp

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/krancour/brignext/v2/apiserver/internal/events"
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

func GetEventsSenderFactoryFromEnvironment() (events.SenderFactory, error) {
	c, err := getConfig()
	if err != nil {
		return nil, err
	}
	return NewEventsSenderFactory(
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
