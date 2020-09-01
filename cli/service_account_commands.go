package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/v2/sdk/api"
	"github.com/krancour/brignext/v2/sdk/meta"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var serviceAccountCommand = &cli.Command{
	Name:  "service-account",
	Usage: "Manage service accounts",
	Subcommands: []*cli.Command{
		{
			Name:  "create",
			Usage: "Create a new service account",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    flagID,
					Aliases: []string{"i"},
					Usage: "Create a service account with the specified ID " +
						"(required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:    flagDescription,
					Aliases: []string{"d"},
					Usage: "Create a service account with the specified " +
						"description (required)",
					Required: true,
				},
			},
			Action: serviceAccountCreate,
		},
		{
			Name:  "get",
			Usage: "Retrieve a service account",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Retrieve the specified service account (required)",
					Required: true,
				},
				cliFlagOutput,
			},
			Action: serviceAccountGet,
		},
		{
			Name:  "grant",
			Usage: "Grant a role to a service account",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Grant a role to the specified service account (required)",
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
			Action: serviceAccountGrant,
		},
		{
			Name:  "list",
			Usage: "Retrieve many service accounts",
			Flags: []cli.Flag{
				cliFlagOutput,
			},
			Action: serviceAccountList,
		},
		{
			Name:  "lock",
			Usage: "Lock a service account out of BrigNext",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Lock the specified service account (required)",
					Required: true,
				},
			},
			Action: serviceAccountLock,
		},
		{
			Name:  "revoke",
			Usage: "Revoke a role from a service account",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Revoke a role to the specified service account (required)",
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
			Action: serviceAccountRevoke,
		},
		{
			Name:  "unlock",
			Usage: "Restore a service account's access to BrigNext",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Unlock the specified service account (required)",
					Required: true,
				},
			},
			Action: serviceAccountUnlock,
		},
	},
}

func serviceAccountCreate(c *cli.Context) error {
	description := c.String(flagDescription)
	id := c.String(flagID)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	token, err := client.ServiceAccounts().Create(
		c.Context,
		api.ServiceAccount{
			ObjectMeta: meta.ObjectMeta{
				ID: id,
			},
			Description: description,
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf("\nService account %q created with token:\n", id)
	fmt.Printf("\n\t%s\n", token.Value)
	fmt.Println(
		"\nStore this token someplace secure NOW. It cannot be retrieved " +
			"later through any other means.",
	)

	return nil
}

func serviceAccountGrant(c *cli.Context) error {
	id := c.String(flagID)
	role := c.String(flagRole)
	scope := c.String(flagScope)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.ServiceAccounts().GrantRole(
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
			"Granted role %q to service account %q.\n\n",
			role,
			id,
		)
	} else {
		fmt.Printf(
			"Granted role %q with scope %q to service account %q.\n\n",
			role,
			scope,
			id,
		)
	}

	return nil
}

func serviceAccountRevoke(c *cli.Context) error {
	id := c.String(flagID)
	role := c.String(flagRole)
	scope := c.String(flagScope)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.ServiceAccounts().RevokeRole(
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
			"Revoked role %q from service account %q.\n\n",
			role,
			id,
		)
	} else {
		fmt.Printf(
			"Revoked role %q with scope %q from service account %q.\n\n",
			role,
			scope,
			id,
		)
	}

	return nil
}

func serviceAccountList(c *cli.Context) error {
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
		serviceAccounts, err :=
			client.ServiceAccounts().List(c.Context, api.ServiceAccountsSelector{}, opts)
		if err != nil {
			return err
		}

		if len(serviceAccounts.Items) == 0 {
			fmt.Println("No service accounts found.")
			return nil
		}

		switch strings.ToLower(output) {
		case "table":
			table := uitable.New()
			table.AddRow("ID", "DESCRIPTION", "AGE", "LOCKED?")
			for _, serviceAccounts := range serviceAccounts.Items {
				table.AddRow(
					serviceAccounts.ID,
					serviceAccounts.Description,
					duration.ShortHumanDuration(time.Since(*serviceAccounts.Created)),
					serviceAccounts.Locked != nil,
				)
			}
			fmt.Println(table)

		case "yaml":
			yamlBytes, err := yaml.Marshal(serviceAccounts)
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from get service accounts operation",
				)
			}
			fmt.Println(string(yamlBytes))

		case "json":
			prettyJSON, err := json.MarshalIndent(serviceAccounts, "", "  ")
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from get service accounts operation",
				)
			}
			fmt.Println(string(prettyJSON))
		}

		if serviceAccounts.RemainingItemCount < 1 ||
			serviceAccounts.Continue == "" {
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
					serviceAccounts.RemainingItemCount,
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

		opts.Continue = serviceAccounts.Continue
	}

	return nil
}

func serviceAccountGet(c *cli.Context) error {
	id := c.String(flagID)
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	serviceAccount, err := client.ServiceAccounts().Get(c.Context, id)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "DESCRIPTION", "AGE", "LOCKED?")
		var age string
		if serviceAccount.Created != nil {
			age = duration.ShortHumanDuration(time.Since(*serviceAccount.Created))
		}
		table.AddRow(
			serviceAccount.ID,
			serviceAccount.Description,
			age,
			serviceAccount.Locked != nil,
		)
		fmt.Println(table)

	case "yaml":
		yamlBytes, err := yaml.Marshal(serviceAccount)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get service account operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case "json":
		prettyJSON, err := json.MarshalIndent(serviceAccount, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get service account operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}

func serviceAccountLock(c *cli.Context) error {
	id := c.String(flagID)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.ServiceAccounts().Lock(c.Context, id); err != nil {
		return err
	}

	fmt.Printf("Service account %q locked.\n", id)

	return nil
}

func serviceAccountUnlock(c *cli.Context) error {
	id := c.String(flagID)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	token, err := client.ServiceAccounts().Unlock(c.Context, id)
	if err != nil {
		return err
	}

	fmt.Printf(
		"\nService account %q unlocked and a new token has been issued:\n",
		id,
	)
	fmt.Printf("\n\t%s\n", token.Value)
	fmt.Println(
		"\nStore this token someplace secure NOW. It cannot be retrieved " +
			"later through any other means.",
	)

	return nil
}
