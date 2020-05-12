package main

import (
	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func getClient(c *cli.Context) (brignext.Client, error) {
	config, err := getConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "error retrieving configuration")
	}
	return brignext.NewClient(
		config.APIAddress,
		config.APIToken,
		c.Bool(flagInsecure),
	), nil
}
