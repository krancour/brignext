package main

import (
	"fmt"
	"io/ioutil"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/common/file"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func eventCreate(c *cli.Context) error {
	payload := c.String(flagPayload)
	payloadFile := c.String(flagPayloadFile)
	projectID := c.String(flagProject)
	source := c.String(flagSource)
	eventType := c.String(flagType)

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
		TypeMeta: brignext.TypeMeta{
			APIVersion: "github.com/krancour/brignext/v2",
			Kind:       "Event",
		},
		EventMeta: brignext.EventMeta{
			ProjectID: projectID,
			Source:    source,
			Type:      eventType,
		},
		Spec: brignext.EventSpec{
			Payload: payload,
		},
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	eventIDs, err := client.CreateEvent(c.Context, event)
	if err != nil {
		return err
	}
	fmt.Printf("Created event %q.\n\n", eventIDs[0])

	return nil
}
