package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func projectDelete(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"project delete requires one argument-- a project ID (case insensitive)",
		)
	}
	id := c.Args()[0]

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.DeleteProject(context.TODO(), id); err != nil {
		return err
	}

	fmt.Printf("Project %q deleted.\n", id)

	return nil
}
