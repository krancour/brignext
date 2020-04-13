package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func eventCancel(c *cli.Context) error {
	// Command-specific flags
	cancelProcessing := c.Bool(flagProcessing)
	projectID := c.String(flagProject)

	// Args
	var eventID string
	if projectID == "" {
		if len(c.Args()) != 1 {
			return errors.New(
				"event cancel requires one argument-- an event ID",
			)
		}
		eventID = c.Args()[0]
	} else if len(c.Args()) != 0 {
		return errors.New(
			"event cancel requires no arguments when the --project flag is used",
		)
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if eventID != "" {
		if canceled, err := client.CancelEvent(
			context.TODO(),
			eventID,
			cancelProcessing,
		); err != nil {
			return err
		} else if canceled {
			fmt.Printf("Event %q canceled.\n", eventID)
		} else {
			return errors.Errorf(
				"event %q was not canceled because specified conditions were not "+
					"satisfied",
				eventID,
			)
		}
	}

	if canceled, err := client.CancelEventsByProject(
		context.TODO(),
		projectID,
		cancelProcessing,
	); err != nil {
		return err
	} else {
		fmt.Printf("Canceled %d events for project %q.\n", canceled, projectID)
	}

	return nil
}
