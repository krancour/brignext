package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func serviceAccountUnlock(c *cli.Context) error {
	id := c.String(flagID)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	token, err := client.ServiceAccounts().Unlock(c.Context, id)
	if err != nil {
		return err
	}

	fmt.Printf(
		"\nService account %q unlocked and a new token has been issued:\n",
		id,
	)
	fmt.Printf("\n\t%s\n", token.Value)
	fmt.Println(
		"\nStore this token someplace secure NOW. It cannot be retrieved " +
			"later through any other means.",
	)

	return nil
}
