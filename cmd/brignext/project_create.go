package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/projects"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func projectCreate(c *cli.Context) error {
	// Inputs
	filename := c.Args()[0]

	// Read and parse the file
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "error reading project file %s", filename)
	}
	project := &brignext.Project{}
	if err := json.Unmarshal(bytes, project); err != nil {
		return errors.Wrapf(err, "error parsing project file %s", filename)
	}

	wireProject := projects.BrignextProjectToWireProject(project)

	// Connect to the API server
	conn, err := getConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := projects.NewProjectsClient(conn)

	// Create the project
	response, err := client.CreateProject(
		context.Background(),
		&projects.CreateProjectRequest{
			Project: wireProject,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error creating project from file %s", filename)
	}

	project = projects.WireProjectToBrignextProject(response.Project)

	// Pretty print the response
	projectJSON, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return errors.Wrap(
			err,
			"error marshaling output from project creation operation",
		)
	}
	fmt.Println(string(projectJSON))

	return nil
}
