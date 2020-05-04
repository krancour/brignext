package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func userUnlock(c *cli.Context) error {
	// Args
	if c.Args().Len() != 1 {
		return errors.New(
			"user unlock requires one argument-- a user ID",
		)
	}
	id := c.Args().Get(0)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.UnlockUser(c.Context, id); err != nil {
		return err
	}

	fmt.Printf("User %q unlocked.\n", id)

	return nil
}
