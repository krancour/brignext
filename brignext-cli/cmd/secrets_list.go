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

func secretsList(c *cli.Context) error {
	// Args
	if c.Args().Len() != 2 {
		return errors.New(
			"secrets list requires two arguments-- a project ID and worker name",
		)
	}
	projectID := c.Args().Get(0)
	workerName := c.Args().Get(1)

	// Command-specific flags
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	secrets, err := client.GetSecrets(c.Context, projectID, workerName)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("KEY", "VALUE")
		for key, value := range secrets {
			table.AddRow(key, value)
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

	return nil
}
