package main

import (
	"fmt"

	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func jobLogs(c *cli.Context) error {
	// Args
	if c.Args().Len() != 3 {
		return errors.New(
			"job logs requires three arguments-- an event ID, a worker name, " +
				"and a job name",
		)
	}
	eventID := c.Args().Get(0)
	workerName := c.Args().Get(1)
	jobName := c.Args().Get(2)

	// Command-specific flags
	follow := c.Bool(flagFollow)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if !follow {
		var logEntries []brignext.LogEntry
		logEntries, err = client.GetJobLogs(c.Context, eventID, workerName, jobName)
		if err != nil {
			return err
		}
		for _, logEntry := range logEntries {
			fmt.Println(logEntry.Message)
		}
		return nil
	}

	logEntryCh, errCh, err := client.StreamJobLogs(
		c.Context,
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
			fmt.Println(logEntry.Message)
		case err := <-errCh:
			return err
		case <-c.Context.Done():
			return nil
		}
	}
}
