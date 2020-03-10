package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func userList(c *cli.Context) error {
	// Args
	if len(c.Args()) != 0 {
		return errors.New("user list requires no arguments")
	}

	// Command-specific flags
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	users, err := client.GetUsers(context.TODO())
	if err != nil {
		return err
	}

	if len(users) == 0 {
		fmt.Println("No users found.")
		return nil
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "NAME", "FIRST SEEN", "LOCKED?")
		for _, user := range users {
			table.AddRow(
				user.ID,
				user.Name,
				user.FirstSeen,
				user.Locked,
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