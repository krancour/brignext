package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/krancour/brignext/pkg/brignext"

	"github.com/gosuri/uitable"
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

	req, err := getRequest(
		http.MethodGet,
		fmt.Sprintf("v2/projects/%s", projectName),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	project := &brignext.Project{}
	if err := json.Unmarshal(respBodyBytes, project); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	if project.Name == "" {
		return errors.Errorf("Project %q not found.", projectName)
	}

	switch output {
	case "table":
		table := uitable.New()
		table.AddRow("NAME", "REPO")
		table.AddRow(
			project.Name,
			project.Repo.Name,
		)
		fmt.Println(table)

	case "json":
		projectJSON, err := json.MarshalIndent(project, "", "  ")
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
