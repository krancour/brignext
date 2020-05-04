package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func serviceAccountLock(c *cli.Context) error {
	// Args
	if c.Args().Len() != 1 {
		return errors.New(
			"service-account lock requires one argument-- a service account ID",
		)
	}
	id := c.Args().Get(0)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.LockServiceAccount(c.Context, id); err != nil {
		return err
	}

	fmt.Printf("Service account %q locked.\n", id)

	return nil
}
