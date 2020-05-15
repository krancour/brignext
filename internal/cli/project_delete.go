package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func projectDelete(c *cli.Context) error {
	id := c.String(flagID)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.DeleteProject(c.Context, id); err != nil {
		return err
	}

	fmt.Printf("Project %q deleted.\n", id)

	return nil
}
