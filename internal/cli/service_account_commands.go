package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/v2"
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
		brignext.NewServiceAccount(id, description),
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

func serviceAccountList(c *cli.Context) error {
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	serviceAccountList, err := client.ServiceAccounts().List(c.Context)
	if err != nil {
		return err
	}

	if len(serviceAccountList.Items) == 0 {
		fmt.Println("No service accounts found.")
		return nil
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "DESCRIPTION", "AGE", "LOCKED?")
		for _, serviceAccount := range serviceAccountList.Items {
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
		}
		fmt.Println(table)

	case "yaml":
		yamlBytes, err := yaml.Marshal(serviceAccountList)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get service accounts operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case "json":
		prettyJSON, err := json.MarshalIndent(serviceAccountList, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get service accounts operation",
			)
		}
		fmt.Println(string(prettyJSON))
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
