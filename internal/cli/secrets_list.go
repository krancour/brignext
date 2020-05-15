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
	output := c.String(flagOutput)
	projectID := c.String(flagProject)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	secretList, err := client.GetSecrets(c.Context, projectID)
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
