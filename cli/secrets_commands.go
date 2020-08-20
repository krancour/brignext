package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/v2/sdk"
	"github.com/krancour/brignext/v2/sdk/meta"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
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

	opts := meta.ListOptions{}

	for {
		secrets, err := client.Projects().ListSecrets(c.Context, projectID, opts)
		if err != nil {
			return err
		}

		switch strings.ToLower(output) {
		case "table":
			table := uitable.New()
			table.AddRow("KEY", "VALUE")
			for _, secret := range secrets.Items {
				table.AddRow(secret.Key, "*** REDACTED ***")
			}
			fmt.Println(table)

		case "yaml":
			yamlBytes, err := yaml.Marshal(secrets)
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from get secrets operation",
				)
			}
			fmt.Println(string(yamlBytes))

		case "json":
			prettyJSON, err := json.MarshalIndent(secrets, "", "  ")
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from get secrets operation",
				)
			}
			fmt.Println(string(prettyJSON))
		}

		if secrets.RemainingItemCount < 1 || secrets.Continue == "" {
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
					secrets.RemainingItemCount,
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

		opts.Continue = secrets.Continue
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

	// Note: The pattern for setting multiple secrets RESTfully in one shot isn't
	// clear, so for now we settle for iterating over the secrets and making an
	// API call for each one. This can be revisited in the future if someone is
	// aware of or discovers the right pattern for this.
	for k, v := range kvPairs {
		secret := sdk.Secret{
			Key:   k,
			Value: v,
		}
		if err := client.Projects().SetSecret(
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

	// Note: The pattern for deleting multiple secrets RESTfully in one shot isn't
	// clear, so for now we settle for iterating over the secrets and making an
	// API call for each one. This can be revisited in the future if someone is
	// aware of or discovers the right pattern for this.
	for _, key := range keys {
		if err := client.Projects().UnsetSecret(
			c.Context,
			projectID,
			key,
		); err != nil {
			return err
		}
	}

	fmt.Printf("Unset secrets for project %q.\n", projectID)

	return nil
}
