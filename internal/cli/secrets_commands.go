package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var secretsCommand = &cli.Command{
	Name:  "secrets",
	Usage: "Manage project secrets",
	Subcommands: []*cli.Command{
		{
			Name:  "list",
			Usage: "Retrieve many secrets; values are always redacted",
			Flags: []cli.Flag{
				cliFlagOutput,
				&cli.StringFlag{
					Name:     flagProject,
					Aliases:  []string{"p"},
					Usage:    "Retrieve secrets for the specified project (required)",
					Required: true,
				},
			},
			Action: secretsList,
		},
		{
			Name:  "set",
			Usage: "Define or redefine the value of one or more secrets",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagProject,
					Aliases:  []string{"p"},
					Usage:    "Set secrets for the specified project (required)",
					Required: true,
				},
				&cli.StringSliceFlag{
					Name:    flagSet,
					Aliases: []string{"s"},
					Usage: "Set a secret using the specified key=value pair " +
						"(required)",
					Required: true,
				},
			},
			Action: secretsSet,
		},
		{
			Name:  "unset",
			Usage: "Clear the value of one or more secrets",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagProject,
					Aliases:  []string{"p"},
					Usage:    "Clear secrets for the specified project",
					Required: true,
				},
				&cli.StringSliceFlag{
					Name:     flagUnset,
					Aliases:  []string{"u"},
					Usage:    "Clear a secret haveing the specified key (required)",
					Required: true,
				},
			},
			Action: secretsUnset,
		},
	},
}

func secretsList(c *cli.Context) error {
	output := c.String(flagOutput)
	projectID := c.String(flagProject)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	secretList, err := client.Secrets().List(c.Context, projectID)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("KEY", "VALUE")
		for _, secret := range secretList.Items {
			table.AddRow(secret.ID, secret.Value)
		}
		fmt.Println(table)

	case "yaml":
		yamlBytes, err := yaml.Marshal(secretList)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get secrets operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case "json":
		prettyJSON, err := json.MarshalIndent(secretList, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get secrets operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}

func secretsSet(c *cli.Context) error {
	projectID := c.String(flagProject)
	kvPairsStr := c.StringSlice(flagSet)

	// We'll make two passes-- we'll parse all the input into a map first,
	// verifying as we go that the input looks good. Only after we know it's good
	// will we iterate over the k/v pairs in the map to set secrets via the API.

	kvPairs := map[string]string{}
	for _, kvPairStr := range kvPairsStr {
		kvTokens := strings.SplitN(kvPairStr, "=", 2)
		if len(kvTokens) != 2 {
			return errors.New("secrets set argument %q is formatted incorrectly")
		}
		kvPairs[kvTokens[0]] = kvTokens[1]
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	// TODO: It would be nicer / more efficient to do a bulk secrets set, but
	// what's the right pattern for doing that restfully?
	for k, v := range kvPairs {
		secret := brignext.Secret{
			TypeMeta: brignext.TypeMeta{
				APIVersion: brignext.APIVersion,
				Kind:       "Secret",
			},
			ObjectMeta: brignext.ObjectMeta{
				ID: k,
			},
			Value: v,
		}
		if err := client.Secrets().Set(
			c.Context,
			projectID,
			secret,
		); err != nil {
			return err
		}
	}

	fmt.Printf("Set secrets for project %q.\n", projectID)

	return nil
}

func secretsUnset(c *cli.Context) error {
	projectID := c.String(flagProject)
	keys := c.StringSlice(flagUnset)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	// TODO: It would be nicer / more efficient to do a bulk secrets unset, but
	// what's the right pattern for doing that restfully?
	for _, secretID := range keys {
		if err := client.Secrets().Unset(
			c.Context,
			projectID,
			secretID,
		); err != nil {
			return err
		}
	}

	fmt.Printf("Unset secrets for project %q.\n", projectID)

	return nil
}
