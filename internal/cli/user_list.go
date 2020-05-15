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

func userList(c *cli.Context) error {
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	userList, err := client.GetUsers(c.Context)
	if err != nil {
		return err
	}

	if len(userList.Items) == 0 {
		fmt.Println("No users found.")
		return nil
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "NAME", "FIRST SEEN", "LOCKED?")
		for _, user := range userList.Items {
			table.AddRow(
				user.ID,
				user.Name,
				user.Created,
				user.Locked != nil,
			)
		}
		fmt.Println(table)

	case "yaml":
		yamlBytes, err := yaml.Marshal(userList)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get users operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case "json":
		prettyJSON, err := json.MarshalIndent(userList, "", "  ")
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
