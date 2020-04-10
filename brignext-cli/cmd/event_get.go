package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/util/duration"
)

func eventGet(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"event get requires one argument-- an event ID",
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

	event, err := client.GetEvent(context.TODO(), id)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "PROJECT", "PROVIDER", "TYPE", "AGE", "PHASE")
		var age string
		if event.Created != nil {
			age = duration.ShortHumanDuration(time.Since(*event.Created))
		}
		table.AddRow(
			event.ID,
			event.ProjectID,
			event.Provider,
			event.Type,
			age,
			event.Status.Phase,
		)
		fmt.Println(table)

		if len(event.Workers) > 0 {
			fmt.Printf("\nEvent %q workers:\n\n", event.ID)
			table = uitable.New()
			table.AddRow("NAME", "STARTED", "ENDED", "PHASE")
			for workerName, worker := range event.Workers {
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
			}
			fmt.Println(table)
		}

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
