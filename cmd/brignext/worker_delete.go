package main

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func workerDelete(c *cli.Context) error {
	// Command-specific flags
	deletePending := c.Bool(flagPending)
	eventID := c.String(flagEvent)
	projectID := c.String(flagProject)
	deleteRunning := c.Bool(flagRunning)

	// Args
	var workerID string
	if projectID == "" && eventID == "" {
		if len(c.Args()) != 1 {
			return errors.New(
				"worker delete requires one argument-- a worker ID",
			)
		}
		workerID = c.Args()[0]
	} else if len(c.Args()) != 0 {
		return errors.New(
			"worker delete requires no arguments when the --project or --event " +
				"flag is used",
		)
	}

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	var path string
	if workerID != "" {
		path = fmt.Sprintf("v2/workers/%s", workerID)
	} else if eventID != "" {
		path = fmt.Sprintf("v2/events/%s/workers", eventID)
	} else {
		path = fmt.Sprintf("v2/projects/%s/workers", projectID)
	}

	req, err := buildRequest(http.MethodDelete, path, nil)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	if deletePending {
		q.Set("pending", "true")
	}
	if deleteRunning {
		q.Set("running", "true")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := getHTTPClient(allowInsecure).Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	if workerID != "" {
		fmt.Printf("Worker %q deleted.\n", workerID)
	} else if eventID != "" {
		fmt.Printf("All workers for event %q deleted.\n", eventID)
	} else {
		fmt.Printf("All workers for project %q deleted.\n", projectID)
	}

	return nil
}
