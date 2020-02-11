package main

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func workerCancel(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"worker cancel requires one argument-- a worker ID",
		)
	}
	id := c.Args()[0]

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	// Command-specific flags
	abortRunning := c.Bool(flagRunning)

	req, err := buildRequest(
		http.MethodPost,
		fmt.Sprintf("v2/workers/%s/stop",
			id,
		), nil)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	if abortRunning {
		q.Set("abortRunning", "true")
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

	fmt.Printf("Worker %q canceled.\n", id)

	return nil
}
