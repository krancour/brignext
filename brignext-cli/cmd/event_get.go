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

func eventGet(c *cli.Context) error {
	// Args
	if c.Args().Len() != 1 {
		return errors.New(
			"event get requires one argument-- an event ID",
		)
	}
	id := c.Args().Get(0)

	// Command-specific flags
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	event, err := client.GetEvent(c.Context, id)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "PROJECT", "SOURCE", "TYPE", "AGE", "WORKER PHASE")
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
			event.Worker.Status.Phase,
		)
		fmt.Println(table)

		if len(event.Worker.Jobs) > 0 {
			fmt.Printf("\nEvent %q worker jobs:\n\n", event.ID)
			table = uitable.New()
			table.AddRow("NAME", "STARTED", "ENDED", "PHASE")
			for jobName, job := range event.Worker.Jobs {
				var started, ended string
				if job.Status.Started != nil {
					started =
						duration.ShortHumanDuration(time.Since(*job.Status.Started))
				}
				if job.Status.Ended != nil {
					ended =
						duration.ShortHumanDuration(time.Since(*job.Status.Ended))
				}
				table.AddRow(
					jobName,
					started,
					ended,
					job.Status.Phase,
				)
			}
			fmt.Println(table)
		}

	case "yaml":
		yamlBytes, err := yaml.Marshal(event)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get event operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case "json":
		prettyJSON, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get event operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}
