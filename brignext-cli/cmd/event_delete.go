package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func eventDelete(c *cli.Context) error {
	// Command-specific flags
	deletePending := c.Bool(flagPending)
	deleteProcessing := c.Bool(flagProcessing)
	projectID := c.String(flagProject)

	// Args
	var eventID string
	if projectID == "" {
		if len(c.Args()) != 1 {
			return errors.New(
				"event delete requires one argument-- an event ID",
			)
		}
		eventID = c.Args()[0]
	} else if len(c.Args()) != 0 {
		return errors.New(
			"event delete requires no arguments when the --project flag is used",
		)
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if eventID != "" {
		var deleted bool
		if deleted, err = client.DeleteEvent(
			context.TODO(),
			eventID,
			deletePending,
			deleteProcessing,
		); err != nil {
			return err
		} else if deleted {
			fmt.Printf("Event %q deleted.\n", eventID)
			return nil
		}
		return errors.Errorf(
			"event %q was not deleted because specified conditions were not "+
				"satisfied",
			eventID,
		)
	}

	deleted, err := client.DeleteEventsByProject(
		context.TODO(),
		projectID,
		deletePending,
		deleteProcessing,
	)
	if err != nil {
		return err
	}
	fmt.Printf("Deleted %d events for project %q.\n", deleted, projectID)

	return nil
}
