package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/pkg/brignext"
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

	req, err := getRequest(http.MethodGet, "v2/projects", nil)
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

	projs := []*brignext.Project{}
	if err := json.Unmarshal(respBodyBytes, &projs); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	if len(projs) == 0 {
		fmt.Println("No projects found.")
		return nil
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
