package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/v2/sdk/api"
	"github.com/krancour/brignext/v2/sdk/meta"
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
			Name:  "grant",
			Usage: "Grant a role to a user",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Grant a role to the specified user (required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:     flagRole,
					Aliases:  []string{"r"},
					Usage:    "Grant the specified role (required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:    flagScope,
					Aliases: []string{"s"},
					Usage: "Constrain the role to the specified scope (required " +
						"for some roles)",
				},
			},
			Action: userGrant,
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
			Name:  "revoke",
			Usage: "Revoke a role from a user",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Revoke a role to the specified user (required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:     flagRole,
					Aliases:  []string{"r"},
					Usage:    "Revoke the specified role (required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:    flagScope,
					Aliases: []string{"s"},
					Usage: "Specify the scope of the role to be revoked (required " +
						"for some roles)",
				},
			},
			Action: userRevoke,
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

	opts := meta.ListOptions{}

	for {
		users, err := client.Users().List(c.Context, api.UsersSelector{}, opts)
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

		// TODO: DRY this up
		var shouldContinue bool
		fmt.Println()
		if err := survey.AskOne(
			&survey.Confirm{
				Message: fmt.Sprintf(
					"%d results remain. Fetch more?",
					users.RemainingItemCount,
				),
			},
			&shouldContinue,
		); err != nil {
			return errors.Wrap(
				err,
				"error confirming if user wishes to continue",
			)
		}
		fmt.Println()
		if !shouldContinue {
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

func userGrant(c *cli.Context) error {
	id := c.String(flagID)
	role := c.String(flagRole)
	scope := c.String(flagScope)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.Users().GrantRole(
		c.Context,
		id,
		api.Role{
			Name:  role,
			Scope: scope,
		},
	); err != nil {
		return err
	}

	if scope == "" {
		fmt.Printf(
			"Granted role %q to user %q.\n\n",
			role,
			id,
		)
	} else {
		fmt.Printf(
			"Granted role %q with scope %q to user %q.\n\n",
			role,
			scope,
			id,
		)
	}

	return nil
}

func userRevoke(c *cli.Context) error {
	id := c.String(flagID)
	role := c.String(flagRole)
	scope := c.String(flagScope)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.Users().RevokeRole(
		c.Context,
		id,
		api.Role{
			Name:  role,
			Scope: scope,
		},
	); err != nil {
		return err
	}

	if scope == "" {
		fmt.Printf(
			"Revoked role %q from user %q.\n\n",
			role,
			id,
		)
	} else {
		fmt.Printf(
			"Revoked role %q with scope %q from user %q.\n\n",
			role,
			scope,
			id,
		)
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
