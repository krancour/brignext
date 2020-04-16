package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func serviceAccountGet(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"service-account get requires one argument-- a service account ID ",
		)
	}
	id := c.Args()[0]

	// Command-specific flags
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	serviceAccount, err := client.GetServiceAccount(context.TODO(), id)
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
			serviceAccount.Locked != nil && *serviceAccount.Locked,
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
