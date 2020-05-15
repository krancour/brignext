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

	// TODO: It would be nicer / more efficient to do a bulk secrets unset, but
	// what's the right pattern for doing that restfully?
	for _, secretID := range keys {
		if err := client.UnsetSecret(
			c.Context,
			projectID,
			secretID,
		); err != nil {
			return err
		}
	}

	fmt.Printf("Unset secrets for project %q.\n", projectID)

	return nil
}
