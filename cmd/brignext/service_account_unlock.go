package main

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func serviceAccountUnlock(c *cli.Context) error {
	// Inputs
	if len(c.Args()) != 1 {
		return errors.New(
			"service-account unlock requires one parameter-- a service account ID",
		)
	}
	id := c.Args()[0]
	allowInsecure := c.GlobalBool(flagInsecure)

	req, err := buildRequest(
		http.MethodDelete,
		fmt.Sprintf("v2/service-accounts/%s/lock", id),
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

	fmt.Printf("Service account %q unlocked.\n", id)

	return nil
}
