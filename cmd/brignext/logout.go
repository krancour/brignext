package main

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func logout(c *cli.Context) error {
	req, err := getRequest(http.MethodDelete, "v2/user/token", nil)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
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
