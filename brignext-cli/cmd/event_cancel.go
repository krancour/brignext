package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func eventCancel(c *cli.Context) error {
	// Command-specific flags
	cancelProcessing := c.Bool(flagProcessing)
	projectID := c.String(flagProject)

	// Args
	var eventID string
	if projectID == "" {
		if c.Args().Len() != 1 {
			return errors.New(
				"event cancel requires one argument-- an event ID",
			)
		}
		eventID = c.Args().Get(0)
	} else if c.Args().Len() != 0 {
		return errors.New(
			"event cancel requires no arguments when the --project flag is used",
		)
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if eventID != "" {
		var canceled bool
		if canceled, err = client.CancelEvent(
			c.Context,
			eventID,
			cancelProcessing,
		); err != nil {
			return err
		} else if canceled {
			fmt.Printf("Event %q canceled.\n", eventID)
			return nil
		}
		return errors.Errorf(
			"event %q was not canceled because specified conditions were not "+
				"satisfied",
			eventID,
		)
	}

	canceled, err := client.CancelEventsByProject(
		c.Context,
		projectID,
		cancelProcessing,
	)
	if err != nil {
		return err
	}
	fmt.Printf("Canceled %d events for project %q.\n", canceled, projectID)

	return nil
}
