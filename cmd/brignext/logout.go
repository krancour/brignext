package main

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func logout(c *cli.Context) error {
	// Args
	if len(c.Args()) != 0 {
		return errors.New("logout requires no arguments")
	}

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	req, err := buildRequest(http.MethodDelete, "v2/session", nil)
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

	if err := deleteConfig(); err != nil {
		return errors.Wrap(err, "error deleting configuration")
	}

	fmt.Println("Logout was successful.")

	return nil
}
