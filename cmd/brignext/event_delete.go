package main

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func eventDelete(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"event delete requires one parameter-- an event ID",
		)
	}
	id := c.Args()[0]

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	// Command-specific flags
	force := c.Bool(flagForce)

	req, err := buildRequest(
		http.MethodDelete,
		fmt.Sprintf("v2/events/%s", id),
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

	fmt.Printf("Event %s deleted.\n", id)

	return nil
}
