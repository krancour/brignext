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

func workerGet(c *cli.Context) error {
	// Args
	if c.Args().Len() != 2 {
		return errors.New(
			"worker get requires two arguments-- an event ID and a worker name",
		)
	}
	eventID := c.Args().Get(0)
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

	worker, err := client.GetWorker(c.Context, eventID, workerName)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("NAME", "STARTED", "ENDED", "PHASE")
		var started, ended string
		if worker.Status.Started != nil {
			started =
				duration.ShortHumanDuration(time.Since(*worker.Status.Started))
		}
		if worker.Status.Ended != nil {
			ended =
				duration.ShortHumanDuration(time.Since(*worker.Status.Ended))
		}
		table.AddRow(
			workerName,
			started,
			ended,
			worker.Status.Phase,
		)
		fmt.Println(table)

		if len(worker.Jobs) > 0 {
			fmt.Printf("\nWorker %q jobs:\n\n", workerName)
			table = uitable.New()
			table.AddRow("NAME", "STARTED", "ENDED", "PHASE")
			for jobName, job := range worker.Jobs {
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
		yamlBytes, err := yaml.Marshal(worker)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get worker operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case "json":
		prettyJSON, err := json.MarshalIndent(worker, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get worker operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}
