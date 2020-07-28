package main

import (
	"context"
	"log"
	"time"

	"github.com/krancour/brignext/v2/internal/events"
	brignext "github.com/krancour/brignext/v2/sdk"
)

func (c *controller) runEventLoop(ctx context.Context, projectID string) {

	var eventReceiver events.Receiver

outerLoop:
	for {

		if eventReceiver != nil {
			closeCtx, cancelCloseCtx :=
				context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelCloseCtx()
			eventReceiver.Close(closeCtx)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		eventReceiver, err := c.eventReceiverFactory.NewReceiver(projectID)
		if err != nil { // It's fatal if we can't get an event receiver
			select {
			case c.errCh <- err:
			case <-ctx.Done():
			}
			return
		}

		// This is the main loop for receiving this project's events
		for {
			// Get the next event for this project
			asyncEvent, err := eventReceiver.Receive(ctx)
			if err != nil {
				continue outerLoop // Try again with a new receiver
			}

			event, err := c.apiClient.Events().Get(ctx, asyncEvent.EventID)
			if err != nil {
				// TODO: We should check what went wrong
				log.Println(err)
				asyncEvent.Ack()
				continue // Next event
			}

			// If the worker's phase isn't PENDING, then there's nothing to do
			if event.Status.WorkerStatus.Phase != brignext.WorkerPhasePending {
				asyncEvent.Ack()
				continue // Next event
			}

			// TODO: Wait for project capacity

			// Wait for system capacity
			select {
			case <-c.availabilityCh:
			case <-ctx.Done():
				// We don't ack the event here because it hasn't been scheduled yet
				continue outerLoop // This will do cleanup before returning
			}

			// Now start the worker...

			if err := c.createWorkspacePVC(ctx, event); err != nil {
				// TODO: Update the event in the database to reflect the error
				log.Printf(
					"error creating workspace for event %q worker: %s",
					asyncEvent.EventID,
					err,
				)
				asyncEvent.Ack()
				continue // Next event
			}

			if err := c.createWorkerPod(ctx, event); err != nil {
				// TODO: Update the event in the database to reflect the error
				log.Printf(
					"error creating pod for event %q worker: %s",
					asyncEvent.EventID,
					err,
				)
				asyncEvent.Ack()
				continue // Next event
			}

		}

	}

}
