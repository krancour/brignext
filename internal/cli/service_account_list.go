package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func serviceAccountList(c *cli.Context) error {
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	serviceAccounts, err := client.GetServiceAccounts(c.Context)
	if err != nil {
		return err
	}

	if len(serviceAccounts) == 0 {
		fmt.Println("No service accounts found.")
		return nil
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "DESCRIPTION", "AGE", "LOCKED?")
		for _, serviceAccount := range serviceAccounts {
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

	return nil
}
