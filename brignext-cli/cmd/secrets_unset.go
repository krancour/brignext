package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func secretsUnset(c *cli.Context) error {
	// Args
	if len(c.Args()) < 2 {
		return errors.New(
			"secrets unset requires at least two arguments-- a project ID, " +
				"a worker name, and a secret key",
		)
	}
	projectID := c.Args()[0]
	workerName := c.Args()[1]
	keys := c.Args()[2:]

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.UnsetSecrets(
		context.TODO(),
		projectID,
		workerName,
		keys,
	); err != nil {
		return err
	}

	fmt.Printf("Unset secrets for project %q.\n", projectID)

	return nil
}
