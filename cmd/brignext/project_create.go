package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func projectCreate(c *cli.Context) error {
	// Inputs
	filename := c.Args()[0]
	allowInsecure := c.GlobalBool(flagInsecure)

	// Read and parse the file
	projectBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "error reading project file %s", filename)
	}

	project := brigade.Project{}
	if err := json.Unmarshal(projectBytes, &project); err != nil {
		return errors.Wrapf(err, "error unmarshaling project file %s", filename)
	}

	req, err := buildRequest(http.MethodPost, "v2/projects", projectBytes)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := getHTTPClient(allowInsecure).Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return errors.Errorf(
			"a project named %q already exists",
			project.Name,
		)
	}
	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	fmt.Printf("Created project %q.\n", project.Name)

	return nil
}
