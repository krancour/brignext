package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func userUnlock(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"user unlock requires one argument-- a user ID",
		)
	}
	id := c.Args()[0]

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.UnlockUser(context.TODO(), id); err != nil {
		return err
	}

	fmt.Printf("User %q unlocked.\n", id)

	return nil
}
