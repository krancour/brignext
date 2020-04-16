package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/util/duration"
)

func projectGet(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"project get requires one argument-- a project ID",
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

	project, err := client.GetProject(context.TODO(), id)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "DESCRIPTION", "AGE")
		var age string
		if project.Created != nil {
			age = duration.ShortHumanDuration(time.Since(*project.Created))
		}
		table.AddRow(
			project.ID,
			project.Description,
			age,
		)
		fmt.Println(table)

	case "yaml":
		yamlBytes, err := yaml.Marshal(project)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get project operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case "json":
		prettyJSON, err := json.MarshalIndent(project, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get project operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}
