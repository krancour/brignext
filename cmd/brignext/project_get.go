package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func projectGet(c *cli.Context) error {
	// Inputs
	if len(c.Args()) != 1 {
		return errors.New(
			"project get requires one parameter-- a project ID",
		)
	}
	id := c.Args()[0]
	output := c.String(flagOutput)
	allowInsecure := c.GlobalBool(flagInsecure)

	switch output {
	case "table":
	case "json":
	default:
		return errors.Errorf("unknown output format %q", output)
	}

	req, err := buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/projects/%s", id),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := getHTTPClient(allowInsecure).Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return errors.Errorf("Project %q not found.", id)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	project := brignext.Project{}
	if err := json.Unmarshal(respBodyBytes, &project); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	switch output {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "DESCRIPTION", "AGE")
		age := "???"
		if project.Created != nil {
			age = duration.ShortHumanDuration(time.Since(*project.Created))
		}
		table.AddRow(
			project.ID,
			project.Description,
			age,
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
