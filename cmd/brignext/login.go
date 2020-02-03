package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func login(c *cli.Context) error {
	// Inputs
	if len(c.Args()) != 1 {
		return errors.New(
			"login requires one parameter-- the address of the API server",
		)
	}
	address := c.Args()[0]
	browseToAuthURL := c.Bool(flagOpen)
	allowInsecure := c.GlobalBool(flagInsecure)

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v2/sessions", address),
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

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	respStruct := struct {
		Token   string `json:"token"`
		AuthURL string `json:"authURL"`
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

	if browseToAuthURL {
		var err error
		switch runtime.GOOS {
		case "linux":
			err = exec.Command("xdg-open", respStruct.AuthURL).Start()
		case "windows":
			err = exec.Command(
				"rundll32",
				"url.dll,FileProtocolHandler",
				respStruct.AuthURL,
			).Start()
		case "darwin":
			err = exec.Command("open", respStruct.AuthURL).Start()
		default:
			err = errors.New("unsupported OS")
		}
		if err != nil {
			return errors.Wrapf(
				err,
				"Error opening authentication URL using the system's default web "+
					"browser.\n\nPlease visit  %s  to complete authentication.\n",
				respStruct.AuthURL,
			)
		}
		return nil
	}

	fmt.Printf(
		"Please visit  %s  to complete authentication.\n",
		respStruct.AuthURL,
	)

	return nil
}
