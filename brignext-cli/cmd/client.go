package main

import (
	"github.com/krancour/brignext/v2/client"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func getClient(c *cli.Context) (client.Client, error) {
	config, err := getConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "error retrieving configuration")
	}
	return client.NewClient(
		config.APIAddress,
		config.APIToken,
		c.GlobalBool(flagInsecure),
	), nil
}
