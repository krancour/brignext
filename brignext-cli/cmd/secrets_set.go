package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func secretsSet(c *cli.Context) error {
	// Args
	if len(c.Args()) < 3 {
		return errors.New(
			"secrets set requires at least three arguments-- a project ID, " +
				"a worker name, and a secret key/value pair delimited by an " +
				"= character",
		)
	}
	projectID := c.Args()[0]
	workerName := c.Args()[1]
	kvPairs := c.Args()[2:]

	secrets := map[string]string{}
	for _, kvPair := range kvPairs {
		kvTokens := strings.SplitN(kvPair, "=", 2)
		if len(kvTokens) != 2 {
			return errors.New("secrets set argument %q is formatted incorrectly")
		}
		secrets[kvTokens[0]] = kvTokens[1]
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.SetSecrets(
		context.TODO(),
		projectID,
		workerName,
		secrets,
	); err != nil {
		return err
	}

	fmt.Printf("Set secrets for project %q.\n", projectID)

	return nil
}
