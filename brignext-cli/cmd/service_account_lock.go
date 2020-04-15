package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func serviceAccountLock(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"service-account lock requires one argument-- a service account ID",
		)
	}
	id := c.Args()[0]

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.LockServiceAccount(context.TODO(), id); err != nil {
		return err
	}

	fmt.Printf("Service account %q locked.\n", id)

	return nil
}
