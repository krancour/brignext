package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/pkg/projects"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func projectGet(c *cli.Context) error {
	// Inputs
	projectName := c.Args()[0]
	output := c.String(flagOutput)

	switch output {
	case "table":
	case "json":
	default:
		return errors.Errorf("unknown output format %q", output)
	}

	// Connect to the API server
	conn, err := getConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := projects.NewProjectsClient(conn)

	// Get the project
	response, err := client.GetProject(
		context.Background(),
		&projects.GetProjectRequest{
			ProjectName: projectName,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error getting project %q", projectName)
	}

	if response.Project == nil {
		return nil
	}

	proj := projects.WireProjectToBrignextProject(response.Project)

	switch output {
	case "table":
		table := uitable.New()
		table.AddRow("NAME", "REPO")
		table.AddRow(
			proj.Name,
			proj.Repo.Name,
		)
		fmt.Println(table)

	case "json":
		projectJSON, err := json.MarshalIndent(proj, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get project operation",
			)
		}
		fmt.Println(string(projectJSON))

	}

	return nil
}
