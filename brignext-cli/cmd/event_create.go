package main

import (
	"context"
	"fmt"

	"github.com/krancour/brignext"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func eventCreate(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"event create requires one argument-- a project ID " +
				"for for which an event should be created",
		)
	}
	projectID := c.Args()[0]

	// Command-specific flags
	eventType := c.String(flagType)
	source := c.String(flagSource)

	event := brignext.Event{
		ProjectID: projectID,
		Source:    source,
		Type:      eventType,
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	eventID, err := client.CreateEvent(context.TODO(), event)
	if err != nil {
		return err
	}

	fmt.Printf("Created event %q.\n\n", eventID)

	// fmt.Println("Streaming event logs...\n")

	// TODO: Stream the logs

	return nil
}
