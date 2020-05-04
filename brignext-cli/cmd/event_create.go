package main

import (
	"fmt"
	"io/ioutil"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/pkg/file"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func eventCreate(c *cli.Context) error {
	// Args
	if c.Args().Len() != 1 {
		return errors.New(
			"event create requires one argument-- a project ID " +
				"for for which an event should be created",
		)
	}
	projectID := c.Args().Get(0)

	// Command-specific flags
	payload := c.String(flagPayload)
	payloadFile := c.String(flagPayloadFile)
	eventType := c.String(flagType)
	source := c.String(flagSource)

	if payload != "" && payloadFile != "" {
		return errors.New(
			"only one of --payload or --payload-file may be specified",
		)
	}
	if payloadFile != "" {
		if !file.Exists(payloadFile) {
			return errors.Errorf("no event payload was found at %s", payloadFile)
		}
		payloadBytes, err := ioutil.ReadFile(payloadFile)
		if err != nil {
			return errors.Wrapf(
				err,
				"error reading event payload from %s",
				payloadFile,
			)
		}
		payload = string(payloadBytes)
	}

	event := brignext.Event{
		ProjectID: projectID,
		Source:    source,
		Type:      eventType,
		Payload:   payload,
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	eventID, err := client.CreateEvent(c.Context, event)
	if err != nil {
		return err
	}

	fmt.Printf("Created event %q.\n\n", eventID)

	// fmt.Println("Streaming event logs...\n")

	// TODO: Stream the logs

	return nil
}
