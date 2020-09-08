package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/brigadecore/brigade/v2/sdk"
)

func getClient(c *cli.Context) (sdk.APIClient, error) {
	config, err := getConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "error retrieving configuration")
	}
	return sdk.NewAPIClient(
		config.APIAddress,
		config.APIToken,
		c.Bool(flagInsecure),
	), nil
}
