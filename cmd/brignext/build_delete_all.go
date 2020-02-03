package main

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func buildDeleteAll(c *cli.Context) error {
	// Inputs
	projectName := c.Args()[0]
	force := c.Bool(flagForce)
	allowInsecure := c.GlobalBool(flagInsecure)

	req, err := buildRequest(
		http.MethodDelete,
		fmt.Sprintf("v2/projects/%s/builds", projectName),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}
	if force {
		req.URL.Query().Set("force", "true")
	}

	resp, err := getHTTPClient(allowInsecure).Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	fmt.Printf("All builds for project %s deleted.\n", projectName)

	return nil
}
