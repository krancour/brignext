package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
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
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "List users",
			Flags: []cli.Flag{
				cliFlagOutput,
			},
			Action: userList,
		},
		{
			Name:  "lock",
			Usage: "Lock a user out of Brigade",
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
			Usage: "Restore a user's access to Brigade",
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
		return err
	}

	opts := meta.ListOptions{}

	for {
		users, err :=
			client.Authx().Users().List(c.Context, nil, &opts)
		if err != nil {
			return err
		}

		if len(users.Items) == 0 {
			fmt.Println("No users found.")
			return nil
		}

		switch strings.ToLower(output) {
		case "table":
			table := uitable.New()
			table.AddRow("ID", "NAME", "FIRST SEEN", "LOCKED?")
			for _, user := range users.Items {
				table.AddRow(
					user.ID,
					user.Name,
					user.Created,
					user.Locked != nil,
				)
			}
			fmt.Println(table)

		case "yaml":
			yamlBytes, err := yaml.Marshal(users)
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from get users operation",
				)
			}
			fmt.Println(string(yamlBytes))

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

		if users.RemainingItemCount < 1 || users.Continue == "" {
			break
		}

		// Exit after one page of output if this isn't a terminal
		if !terminal.IsTerminal(int(os.Stdout.Fd())) {
			break
		}

		if shouldContinue, err :=
			shouldContinue(users.RemainingItemCount); err != nil {
			return err
		} else if !shouldContinue {
			break
		}

		opts.Continue = users.Continue
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
		return err
	}

	user, err := client.Authx().Users().Get(c.Context, id)
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
		return err
	}

	if err := client.Authx().Users().Lock(c.Context, id); err != nil {
		return err
	}

	fmt.Printf("User %q locked.\n", id)

	return nil
}

func userUnlock(c *cli.Context) error {
	id := c.String(flagID)

	client, err := getClient(c)
	if err != nil {
		return err
	}

	if err := client.Authx().Users().Unlock(c.Context, id); err != nil {
		return err
	}

	fmt.Printf("User %q unlocked.\n", id)

	return nil
}
