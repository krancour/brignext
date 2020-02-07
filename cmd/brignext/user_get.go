package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func userGet(c *cli.Context) error {
	// Inputs
	if len(c.Args()) != 1 {
		return errors.New(
			"user get requires one parameter-- a username",
		)
	}
	username := c.Args()[0]
	output := c.String(flagOutput)
	allowInsecure := c.GlobalBool(flagInsecure)

	switch output {
	case "table":
	case "json":
	default:
		return errors.Errorf("unknown output format %q", output)
	}

	req, err := buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/users/%s", username),
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

	user := struct {
		Username  string    `json:"username"`
		FirstSeen time.Time `json:"firstSeen"`
	}{}
	if err := json.Unmarshal(respBodyBytes, &user); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	if user.Username == "" {
		return errors.Errorf("User %q not found.", username)
	}

	switch output {
	case "table":
		table := uitable.New()
		table.AddRow("USERNAME", "FIRST SEEN")
		table.AddRow(
			user.Username,
			user.FirstSeen,
		)
		fmt.Println(table)

	case "json":
		projectJSON, err := json.MarshalIndent(user, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get project operation",
			)
		}
		fmt.Println(string(projectJSON))
	}

	return nil
}
