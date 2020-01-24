package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/projects"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func projectList(c *cli.Context) error {
	// Inputs
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

	// Get the projects
	response, err := client.GetProjects(
		context.Background(),
		&projects.GetProjectsRequest{},
	)
	if err != nil {
		return errors.Wrap(err, "error listing projects")
	}

	projs := make([]*brignext.Project, len(response.Projects))
	for i, wireProject := range response.Projects {
		projs[i] = projects.WireProjectToBrignextProject(wireProject)
	}

	switch output {
	case "table":
		table := uitable.New()
		table.AddRow("NAME", "REPO")
		for _, project := range projs {
			table.AddRow(
				project.Name,
				project.Repo.Name,
			)
		}
		fmt.Println(table)

	case "json":
		responseJSON, err := json.MarshalIndent(projs, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get projects operation",
			)
		}
		fmt.Println(string(responseJSON))

	}

	return nil
}
