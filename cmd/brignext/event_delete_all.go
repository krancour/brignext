package main

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func eventDeleteAll(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"event delete-all requires one parameter-- a project ID " +
				"(case insensitive)",
		)
	}
	projectID := c.Args()[0]

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	// Command-specific flags
	force := c.Bool(flagForce)

	req, err := buildRequest(
		http.MethodDelete,
		fmt.Sprintf("v2/projects/%s/events", projectID),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}
	if force {
		q := req.URL.Query()
		q.Set("force", "true")
		req.URL.RawQuery = q.Encode()
	}

	resp, err := getHTTPClient(allowInsecure).Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	fmt.Printf("All events for project %q deleted.\n", projectID)

	return nil
}
