package main

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func eventCancel(c *cli.Context) error {
	// Command-specific flags
	cancelProcessing := c.Bool(flagProcessing)
	projectID := c.String(flagProject)

	// Args
	var eventID string
	if projectID == "" {
		if len(c.Args()) != 1 {
			return errors.New(
				"event cancel requires one argument-- an event ID",
			)
		}
		eventID = c.Args()[0]
	} else if len(c.Args()) != 0 {
		return errors.New(
			"event cancel requires no arguments when the --project flag is used",
		)
	}

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	var path string
	if eventID != "" {
		path = fmt.Sprintf("v2/events/%s/stop", eventID)
	} else {
		path = fmt.Sprintf("v2/projects/%s/events/stop", projectID)
	}

	req, err := buildRequest(http.MethodPost, path, nil)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	if cancelProcessing {
		q.Set("abortProcessing", "true")
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

	if eventID != "" {
		fmt.Printf("Event %q canceled.\n", eventID)
	} else {
		fmt.Printf("All events for project %q canceled.\n", projectID)
	}

	return nil
}
