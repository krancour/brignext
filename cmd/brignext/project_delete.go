package main

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func projectDelete(c *cli.Context) error {
	// Inputs
	projectName := c.Args()[0]
	allowInsecure := c.GlobalBool(flagInsecure)

	req, err := buildRequest(
		http.MethodDelete,
		fmt.Sprintf("v2/projects/%s", projectName),
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

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	fmt.Printf("Project %q deleted.\n", projectName)

	return nil
}
