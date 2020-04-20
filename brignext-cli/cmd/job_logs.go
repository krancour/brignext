package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func jobLogs(c *cli.Context) error {
	ctx := context.TODO()

	// Args
	if len(c.Args()) != 3 {
		return errors.New(
			"job logs requires three arguments-- an event ID, a worker name, " +
				"and a job name",
		)
	}
	eventID := c.Args()[0]
	workerName := c.Args()[1]
	jobName := c.Args()[2]

	// Command-specific flags
	follow := c.Bool(flagFollow)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if !follow {
		logEntries, err := client.GetJobLogs(ctx, eventID, workerName, jobName)
		if err != nil {
			return err
		}
		for _, logEntry := range logEntries {
			fmt.Print(logEntry.Message)
		}
		return nil
	}

	logEntryCh, errCh, err := client.StreamJobLogs(
		ctx,
		eventID,
		workerName,
		jobName,
	)
	if err != nil {
		return err
	}
	for {
		select {
		case logEntry := <-logEntryCh:
			fmt.Print(logEntry.Message)
		case err := <-errCh:
			return err
		case <-ctx.Done():
			return nil
		}
	}
}
