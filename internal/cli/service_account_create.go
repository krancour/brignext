package main

import (
	"fmt"

	"github.com/krancour/brignext/v2"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func serviceAccountCreate(c *cli.Context) error {
	description := c.String(flagDescription)
	id := c.String(flagID)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	token, err := client.CreateServiceAccount(
		c.Context,
		brignext.ServiceAccount{
			TypeMeta: brignext.TypeMeta{
				APIVersion: "github.com/krancour/brignext/v2",
				Kind:       "ServiceAccount",
			},
			ServiceAccountMeta: brignext.ServiceAccountMeta{
				ID:          id,
				Description: description,
			},
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf("\nService account %q created with token:\n", id)
	fmt.Printf("\n\t%s\n", token)
	fmt.Println(
		"\nStore this token someplace secure NOW. It cannot be retrieved " +
			"later through any other means.",
	)

	return nil
}
