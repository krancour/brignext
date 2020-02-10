package main

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func serviceAccountLock(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"service-account lock requires one parameter-- a service account ID " +
				"(case insensitive)",
		)
	}
	id := c.Args()[0]

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	req, err := buildRequest(
		http.MethodPost,
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

	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	fmt.Printf("Service account %q locked.\n", id)

	return nil
}
