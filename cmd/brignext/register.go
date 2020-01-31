package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func register(c *cli.Context) error {
	// Inputs
	address := c.Args()[0]
	username := c.String(flagUsername)
	password := c.String(flagPassword)

	// TODO: There should be an option to get username and password interactively
	// if not specified, otherwise username and password could show up in shell
	// history, which users may not want in some cases.

	requestBody, err := json.Marshal(
		struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: username,
			Password: password,
		},
	)
	if err != nil {
		return errors.Wrap(err, "error marshaling request body")
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v2/users", address),
		bytes.NewBuffer(requestBody),
	)
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

	fmt.Println("Registration was successful.")

	return doLogin(address, username, password)
}
