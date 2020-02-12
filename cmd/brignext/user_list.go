package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func userList(c *cli.Context) error {
	// Args
	if len(c.Args()) != 0 {
		return errors.New("user list requires no arguments")
	}

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	// Command-specific flags
	output := c.String(flagOutput)

	switch output {
	case "table":
	case "json":
	default:
		return errors.Errorf("unknown output format %q", output)
	}

	req, err := buildRequest(http.MethodGet, "v2/users", nil)
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

	users := []brignext.User{}
	if err := json.Unmarshal(respBodyBytes, &users); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	if len(users) == 0 {
		fmt.Println("No users found.")
		return nil
	}

	switch output {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "NAME", "FIRST SEEN", "LOCKED?")
		for _, user := range users {
			table.AddRow(
				user.ID,
				user.Name,
				user.FirstSeen,
				user.Locked != nil && *user.Locked,
			)
		}
		fmt.Println(table)

	case "json":
		prettyJSON, err := json.MarshalIndent(users, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get users operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}
