package main

import (
	"context"
	"fmt"

	"github.com/krancour/brignext/pkg/builds"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func buildDeleteAll(c *cli.Context) error {
	// Inputs
	projectName := c.Args()[0]
	force := c.Bool(flagForce)

	conn, err := getConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := builds.NewBuildsClient(conn)
	if _, err = client.DeleteAllBuilds(
		context.Background(),
		&builds.DeleteAllBuildsRequest{
			ProjectName: projectName,
			Force:       force,
		},
	); err != nil {
		return errors.Wrap(err, "error deleting all builds")
	}

	fmt.Printf("All builds for project %s deleted.\n", projectName)

	return nil
}
