package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func userGet(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"user get requires one argument-- a user ID (case insensitive)",
		)
	}
	id := c.Args()[0]

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	// Command-specific flags
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	req, err := buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/users/%s", id),
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

	if resp.StatusCode == http.StatusNotFound {
		return errors.Errorf("User %q not found.", id)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	user := brignext.User{}
	if err := json.Unmarshal(respBodyBytes, &user); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "NAME", "FIRST SEEN", "LOCKED?")
		table.AddRow(
			user.ID,
			user.Name,
			user.FirstSeen,
			user.Locked != nil && *user.Locked,
		)
		fmt.Println(table)

	case "json":
		prettyJSON, err := json.MarshalIndent(user, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get user operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}
