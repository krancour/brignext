package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func secretsUnset(c *cli.Context) error {
	projectID := c.String(flagProject)
	keys := c.StringSlice(flagUnset)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.UnsetSecrets(
		c.Context,
		projectID,
		keys,
	); err != nil {
		return err
	}

	fmt.Printf("Unset secrets for project %q.\n", projectID)

	return nil
}
