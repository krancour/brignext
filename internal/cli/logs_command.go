package main

import (
	"fmt"

	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var logsCommand = &cli.Command{
	Name:  "logs",
	Usage: "View worker or job logs",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     flagEvent,
			Aliases:  []string{"e"},
			Usage:    "View logs from the specified event",
			Required: true,
		},
		&cli.BoolFlag{
			Name:    flagFollow,
			Aliases: []string{"f"},
			Usage:   "If set, will stream logs until interrupted",
		},
		&cli.BoolFlag{
			Name:    flagInit,
			Aliases: []string{"i"},
			Usage:   "View logs from the corresponding init container",
		},
		&cli.StringFlag{
			Name:    flagJob,
			Aliases: []string{"j"},
			Usage: "View logs from the specified job; if not set, displays " +
				"worker logs",
		},
	},
	Action: logs,
}

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
