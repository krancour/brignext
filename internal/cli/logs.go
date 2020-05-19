package main

import (
	"fmt"

	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func logs(c *cli.Context) error {
	eventID := c.String(flagEvent)
	follow := c.Bool(flagFollow)
	// initLogs := c.Bool(flagInits)
	jobName := c.String(flagJob)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if !follow {
		var logEntryList brignext.LogEntryList
		if jobName == "" {
			logEntryList, err = client.Events().GetWorkerLogs(c.Context, eventID)
		} else {
			logEntryList, err = client.Events().GetJobLogs(
				c.Context,
				eventID,
				jobName,
			)
		}
		if err != nil {
			return err
		}
		for _, logEntry := range logEntryList.Items {
			fmt.Println(logEntry.Message)
		}
		return nil
	}

	var logEntryCh <-chan brignext.LogEntry
	var errCh <-chan error
	if jobName == "" {
		logEntryCh, errCh, err = client.Events().StreamWorkerLogs(
			c.Context,
			eventID,
		)
	} else {
		logEntryCh, errCh, err = client.Events().StreamJobLogs(
			c.Context,
			eventID,
			jobName,
		)
	}
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
