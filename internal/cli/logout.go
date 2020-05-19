package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func logout(c *cli.Context) error {
	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	// We're ignoring any error here because even if the session wasn't found
	// and deleted server-side, we still want to move on to destroying the local
	// token.
	client.Sessions().Delete(c.Context) // nolint: errcheck

	if err := deleteConfig(); err != nil {
		return errors.Wrap(err, "error deleting configuration")
	}

	fmt.Println("Logout was successful.")

	return nil
}