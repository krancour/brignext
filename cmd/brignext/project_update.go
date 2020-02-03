package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func projectUpdate(c *cli.Context) error {
	// Inputs
	filename := c.Args()[0]
	allowInsecure := c.GlobalBool(flagInsecure)

	// Read and parse the file
	projectBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "error reading project file %s", filename)
	}

	project := &brignext.Project{}
	if err := json.Unmarshal(projectBytes, project); err != nil {
		return errors.Wrapf(err, "error parsing project file %s", filename)
	}

	req, err := buildRequest(
		http.MethodPut,
		fmt.Sprintf("v2/projects/%s", project.Name),
		projectBytes,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := getHTTPClient(allowInsecure).Do(req)
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

	project = &brignext.Project{}
	if err := json.Unmarshal(respBodyBytes, project); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

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
