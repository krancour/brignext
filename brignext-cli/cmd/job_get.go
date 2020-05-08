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

func jobGet(c *cli.Context) error {
	// Args
	if c.Args().Len() != 2 {
		return errors.New(
			"job get requires two arguments-- an event ID and a job name",
		)
	}
	eventID := c.Args().Get(0)
	jobName := c.Args().Get(2)

	// Command-specific flags
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	job, err := client.GetJob(c.Context, eventID, jobName)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("NAME", "STARTED", "ENDED", "PHASE")
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
		fmt.Println(table)

	case "yaml":
		yamlBytes, err := yaml.Marshal(job)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get job operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case "json":
		prettyJSON, err := json.MarshalIndent(job, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get job operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}
