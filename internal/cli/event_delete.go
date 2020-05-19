package main

import (
	"fmt"

	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func eventDelete(c *cli.Context) error {
	eventID := c.String(flagID)
	deletePending := c.Bool(flagPending)
	projectID := c.String(flagProject)
	deleteRunning := c.Bool(flagRunning)

	if eventID == "" && projectID == "" {
		return errors.New("one of --id or --project must be set")
	}

	if eventID != "" && projectID != "" {
		return errors.New("--id and --project are mutually exclusive")
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if eventID != "" {
		var eventRefList brignext.EventReferenceList
		if eventRefList, err = client.Events().Delete(
			c.Context,
			eventID,
			deletePending,
			deleteRunning,
		); err != nil {
			return err
		}
		if len(eventRefList.Items) != 0 {
			fmt.Printf("Event %q deleted.\n", eventID)
			return nil
		}
		return errors.Errorf(
			"event %q was not deleted because specified conditions were not "+
				"satisfied",
			eventID,
		)
	}

	eventRefList, err := client.Events().DeleteByProject(
		c.Context,
		projectID,
		deletePending,
		deleteRunning,
	)
	if err != nil {
		return err
	}
	fmt.Printf(
		"Deleted %d events for project %q.\n",
		len(eventRefList.Items),
		projectID,
	)

	return nil
}
