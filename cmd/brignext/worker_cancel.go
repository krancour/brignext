package main

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func workerCancel(c *cli.Context) error {
	// Command-specific flags
	cancelPending := c.Bool(flagPending)
	eventID := c.String(flagEvent)
	projectID := c.String(flagProject)
	cancelRunning := c.Bool(flagRunning)

	// Args
	var workerID string
	if projectID == "" && eventID == "" {
		if len(c.Args()) != 1 {
			return errors.New(
				"worker cancel requires one argument-- a worker ID",
			)
		}
		workerID = c.Args()[0]
	} else if len(c.Args()) != 0 {
		return errors.New(
			"worker cancel requires no arguments when the --project or --event " +
				"flag is used",
		)
	}

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	var path string
	if workerID != "" {
		path = fmt.Sprintf("v2/workers/%s/stop", workerID)
	} else if eventID != "" {
		path = fmt.Sprintf("v2/events/%s/workers/stop", eventID)
	} else {
		path = fmt.Sprintf("v2/projects/%s/workers/stop", projectID)
	}

	req, err := buildRequest(http.MethodPost, path, nil)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	if cancelPending {
		q.Set("pending", "true")
	}
	if cancelRunning {
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
		fmt.Printf("Worker %q canceled.\n", workerID)
	} else if eventID != "" {
		fmt.Printf("All workers for event %q canceled.\n", eventID)
	} else {
		fmt.Printf("All workers for project %q canceled.\n", projectID)
	}

	return nil
}
