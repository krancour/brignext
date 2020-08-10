package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/v2/sdk/api"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var userCommand = &cli.Command{
	Name:  "user",
	Usage: "Manage users",
	Subcommands: []*cli.Command{
		{
			Name:  "get",
			Usage: "Retrieve a user",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Retrieve the specified user (required)",
					Required: true,
				},
				cliFlagOutput,
			},
			Action: userGet,
		},
		{
			Name:  "list",
			Usage: "Retrieve many users",
			Flags: []cli.Flag{
				cliFlagOutput,
			},
			Action: userList,
		},
		{
			Name:  "lock",
			Usage: "Lock a user out of BrigNext",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Lock the specified user (required)",
					Required: true,
				},
			},
			Action: userLock,
		},
		{
			Name:  "unlock",
			Usage: "Restore a user's access to BrigNext",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Unlock the specified user (required)",
					Required: true,
				},
			},
			Action: userUnlock,
		},
	},
}

func userList(c *cli.Context) error {
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	userList, err := client.Users().List(c.Context, api.UserListOptions{})
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

	user, err := client.Users().Get(c.Context, id)
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
			user.Created,
			user.Locked != nil,
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

func userLock(c *cli.Context) error {
	id := c.String(flagID)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.Users().Lock(c.Context, id); err != nil {
		return err
	}

	fmt.Printf("User %q locked.\n", id)

	return nil
}

func userUnlock(c *cli.Context) error {
	id := c.String(flagID)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.Users().Unlock(c.Context, id); err != nil {
		return err
	}

	fmt.Printf("User %q unlocked.\n", id)

	return nil
}
