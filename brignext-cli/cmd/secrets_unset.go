package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func secretsUnset(c *cli.Context) error {
	// Args
	if c.Args().Len() < 2 {
		return errors.New(
			"secrets unset requires at least two arguments-- a project ID, " +
				"a worker name, and a secret key",
		)
	}
	projectID := c.Args().Get(0)
	workerName := c.Args().Get(1)
	keys := c.Args().Slice()[2:]

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.UnsetSecrets(
		c.Context,
		projectID,
		workerName,
		keys,
	); err != nil {
		return err
	}

	fmt.Printf("Unset secrets for project %q.\n", projectID)

	return nil
}
