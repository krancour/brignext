package main

import (
	"context"
	"fmt"

	"github.com/krancour/brignext/pkg/projects"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func projectDelete(c *cli.Context) error {
	// Inputs
	projectName := c.Args()[0]

	// Connect to the API server
	conn, err := getConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := projects.NewProjectsClient(conn)

	// Delete the project
	if _, err = client.DeleteProject(
		context.Background(),
		&projects.DeleteProjectRequest{
			ProjectName: projectName,
		},
	); err != nil {
		return errors.Wrap(err, "error deleting project")
	}

	fmt.Printf("Project %q deleted.\n", projectName)

	return nil
}
