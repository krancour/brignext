package main

import (
	"fmt"
	"strings"

	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func secretsSet(c *cli.Context) error {
	projectID := c.String(flagProject)
	kvPairsStr := c.StringSlice(flagSet)

	// We'll make two passes-- we'll parse all the input into a map first,
	// verifying as we go that the input looks good. Only after we know it's good
	// will we iterate over the k/v pairs in the map to set secrets via the API.

	kvPairs := map[string]string{}
	for _, kvPairStr := range kvPairsStr {
		kvTokens := strings.SplitN(kvPairStr, "=", 2)
		if len(kvTokens) != 2 {
			return errors.New("secrets set argument %q is formatted incorrectly")
		}
		kvPairs[kvTokens[0]] = kvTokens[1]
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	// TODO: It would be nicer / more efficient to do a bulk secrets set, but
	// what's the right pattern for doing that restfully?
	for k, v := range kvPairs {
		secret := brignext.Secret{
			TypeMeta: brignext.TypeMeta{
				APIVersion: brignext.APIVersion,
				Kind:       "Secret",
			},
			ObjectMeta: brignext.ObjectMeta{
				ID: k,
			},
			Value: v,
		}
		if err := client.Secrets().Set(
			c.Context,
			projectID,
			secret,
		); err != nil {
			return err
		}
	}

	fmt.Printf("Set secrets for project %q.\n", projectID)

	return nil
}
