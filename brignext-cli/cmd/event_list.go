package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"k8s.io/apimachinery/pkg/util/duration"
)

func eventList(c *cli.Context) error {
	// Args
	if c.Args().Len() != 0 {
		return errors.New(
			"event list requires no arguments",
		)
	}

	// Command-specific flags
	output := c.String(flagOutput)
	projectID := c.String(flagProject)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	var events []brignext.Event
	if projectID == "" {
		events, err = client.GetEvents(c.Context)
	} else {
		events, err = client.GetEventsByProject(c.Context, projectID)
	}
	if err != nil {
		return err
	}

	if len(events) == 0 {
		fmt.Println("No events found.")
		return nil
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "PROJECT", "SOURCE", "TYPE", "AGE", "PHASE")
		for _, event := range events {
			var age string
			if event.Created != nil {
				age = duration.ShortHumanDuration(time.Since(*event.Created))
			}
			table.AddRow(
				event.ID,
				event.ProjectID,
				event.Source,
				event.Type,
				age,
				event.Status.Phase,
			)
		}
		fmt.Println(table)

	case "yaml":
		yamlBytes, err := yaml.Marshal(events)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get events operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case "json":
		prettyJSON, err := json.MarshalIndent(events, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get events operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}
