package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func workerCancel(c *cli.Context) error {
	// Command-specific flags
	cancelRunning := c.Bool(flagRunning)

	// Args
	if c.Args().Len() != 2 {
		return errors.New(
			"worker cancel requires two arguments-- thr event ID and worker name",
		)
	}
	eventID := c.Args().Get(0)
	workerName := c.Args().Get(1)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if canceled, err := client.CancelWorker(
		c.Context,
		eventID,
		workerName,
		cancelRunning,
	); err != nil {
		return err
	} else if canceled {
		fmt.Printf("Event %q worker %q canceled.\n", eventID, workerName)
	} else {
		return errors.Errorf(
			"event %q workjer %q was not canceled because specified conditions "+
				"were not satisfied",
			eventID,
			workerName,
		)
	}

	return nil
}
