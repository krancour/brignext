package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func serviceAccountUnlock(c *cli.Context) error {
	// Args
	if c.Args().Len() != 1 {
		return errors.New(
			"service-account unlock requires one argument-- a service account ID",
		)
	}
	id := c.Args().Get(0)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	token, err := client.UnlockServiceAccount(c.Context, id)
	if err != nil {
		return err
	}

	fmt.Printf(
		"\nService account %q unlocked and a new token has been issued:\n",
		id,
	)
	fmt.Printf("\n\t%s\n", token)
	fmt.Println(
		"\nStore this token someplace secure NOW. It cannot be retrieved " +
			"later through any other means.",
	)

	return nil
}
