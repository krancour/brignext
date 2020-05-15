package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"k8s.io/apimachinery/pkg/util/duration"
)

func projectList(c *cli.Context) error {
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	projects, err := client.GetProjects(c.Context)
	if err != nil {
		return err
	}

	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return nil
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "DESCRIPTION", "AGE")
		for _, project := range projects {
			var age string
			if project.Created != nil {
				age = duration.ShortHumanDuration(time.Since(*project.Created))
			}
			table.AddRow(
				project.ID,
				project.Spec.Description,
				age,
			)
		}
		fmt.Println(table)

	case "yaml":
		yamlBytes, err := yaml.Marshal(projects)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get projects operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case "json":
		prettyJSON, err := json.MarshalIndent(projects, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get projects operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}
