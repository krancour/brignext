package main

import (
	"fmt"

	"github.com/krancour/brignext/v2/sdk"
	"github.com/krancour/brignext/v2/sdk/api"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var logsCommand = &cli.Command{
	Name:  "logs",
	Usage: "View worker or job logs",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    flagContainer,
			Aliases: []string{"c"},
			Usage: "View logs from the specified container; if not set, displays " +
				"logs from the worker or job's \"primary\" container",
		},
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

	opts := api.LogOptions{
		Job:       c.String(flagJob),
		Container: c.String(flagContainer),
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if !follow {
		var logEntryList sdk.LogEntryList
		if logEntryList, err =
			client.Events().GetLogs(c.Context, eventID, opts); err != nil {
			return err
		}
		for _, logEntry := range logEntryList.Items {
			fmt.Println(logEntry.Message)
		}
		return nil
	}

	logEntryCh, errCh, err := client.Events().StreamLogs(c.Context, eventID, opts)
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
