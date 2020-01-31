package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func login(c *cli.Context) error {
	// Inputs
	address := c.Args()[0]
	username := c.String(flagUsername)
	password := c.String(flagPassword)

	// TODO: There should be an option to get username and password interactively
	// if not specified, otherwise username and password could show up in shell
	// history, which users may not want in some cases.

	// TODO: Log out of any API server we're already logged into

	return doLogin(address, username, password)
}

func doLogin(address, username, password string) error {
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v2/users/%s/tokens", address, username),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}
	req.SetBasicAuth(username, password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	respStruct := struct {
		Token string `json:"token"`
	}{}
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	if err := saveConfig(
		&config{
			APIAddress: address,
			APIToken:   respStruct.Token,
		},
	); err != nil {
		return errors.Wrap(err, "error persisting configuration")
	}

	fmt.Println("Login was successful.")

	return nil
}
