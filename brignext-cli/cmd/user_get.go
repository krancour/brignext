package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func userGet(c *cli.Context) error {
	id := c.String(flagID)
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	user, err := client.GetUser(c.Context, id)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "NAME", "FIRST SEEN", "LOCKED?")
		table.AddRow(
			user.ID,
			user.Name,
			user.FirstSeen,
			user.Locked,
		)
		fmt.Println(table)

	case "yaml":
		yamlBytes, err := yaml.Marshal(user)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get user operation",
			)
		}
		fmt.Println(string(yamlBytes))

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
