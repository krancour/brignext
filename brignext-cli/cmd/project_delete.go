package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func projectDelete(c *cli.Context) error {
	// Args
	if c.Args().Len() != 1 {
		return errors.New(
			"project delete requires one argument-- a project ID",
		)
	}
	id := c.Args().Get(0)

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
