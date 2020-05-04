package main

import (
	"context"
	"fmt"

	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func workerLogs(c *cli.Context) error {
	ctx := context.TODO()

	// Args
	if len(c.Args()) != 2 {
		return errors.New(
			"worker logs requires two arguments-- an event ID and a worker name",
		)
	}
	eventID := c.Args()[0]
	workerName := c.Args()[1]

	// Command-specific flags
	follow := c.Bool(flagFollow)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if !follow {
		var logEntries []brignext.LogEntry
		logEntries, err = client.GetWorkerLogs(ctx, eventID, workerName)
		if err != nil {
			return err
		}
		for _, logEntry := range logEntries {
			fmt.Print(logEntry.Message)
		}
		return nil
	}

	logEntryCh, errCh, err := client.StreamWorkerLogs(ctx, eventID, workerName)
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
