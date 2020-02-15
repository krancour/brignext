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

func projectCreate(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"project create requires one argument-- a path to a file containing a " +
				"project definition",
		)
	}
	filename := c.Args()[0]

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	// Read and parse the file
	projectBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "error reading project file %s", filename)
	}

	project := brignext.Project{}
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
			"a project with the ID %q already exists",
			project.ID,
		)
	}
	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	fmt.Printf("Created project %q.\n", project.ID)

	return nil
}
