package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func secretsSet(c *cli.Context) error {
	projectID := c.String(flagProject)
	kvPairs := c.StringSlice(flagSet)

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
		c.Context,
		projectID,
		secrets,
	); err != nil {
		return err
	}

	fmt.Printf("Set secrets for project %q.\n", projectID)

	return nil
}
