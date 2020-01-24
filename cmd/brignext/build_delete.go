package main

import (
	"context"
	"fmt"

	"github.com/krancour/brignext/pkg/builds"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func buildDelete(c *cli.Context) error {
	// Inputs
	id := c.Args()[0]
	force := c.Bool(flagForce)

	conn, err := getConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := builds.NewBuildsClient(conn)
	if _, err = client.DeleteBuild(
		context.Background(),
		&builds.DeleteBuildRequest{
			Id:    id,
			Force: force,
		},
	); err != nil {
		return errors.Wrap(err, "error deleting build")
	}

	fmt.Printf("Build %s deleted.\n", id)

	return nil
}
