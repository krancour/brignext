package main

import (
	"github.com/brigadecore/brigade/v2/sdk/api"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func getClient(c *cli.Context) (api.Client, error) {
	config, err := getConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "error retrieving configuration")
	}
	return api.NewClient(
		config.APIAddress,
		config.APIToken,
		c.Bool(flagInsecure),
	), nil
}
