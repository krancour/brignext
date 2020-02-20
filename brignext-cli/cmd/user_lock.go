package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func userLock(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"user lock requires one argument-- a user ID (case insensitive)",
		)
	}
	id := c.Args()[0]

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.LockUser(context.TODO(), id); err != nil {
		return err
	}

	fmt.Printf("User %q locked.\n", id)

	return nil
}
