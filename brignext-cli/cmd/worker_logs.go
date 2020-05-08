package main

import (
	"fmt"

	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func workerLogs(c *cli.Context) error {
	// Args
	if c.Args().Len() != 1 {
		return errors.New(
			"worker logs requires one arguments-- an event ID",
		)
	}
	eventID := c.Args().Get(0)

	// Command-specific flags
	follow := c.Bool(flagFollow)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if !follow {
		var logEntries []brignext.LogEntry
		logEntries, err = client.GetWorkerLogs(c.Context, eventID)
		if err != nil {
			return err
		}
		for _, logEntry := range logEntries {
			fmt.Println(logEntry.Message)
		}
		return nil
	}

	logEntryCh, errCh, err := client.StreamWorkerLogs(
		c.Context,
		eventID,
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
