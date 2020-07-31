package main

import (
	"context"
	"log"
	"time"

	"github.com/krancour/brignext/v2/scheduler/internal/events"
	"github.com/krancour/brignext/v2/sdk"
)

func (s *scheduler) runEventLoop(ctx context.Context, projectID string) {

	var eventsReceiver events.Receiver

outerLoop:
	for {

		if eventsReceiver != nil {
			closeCtx, cancelCloseCtx :=
				context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelCloseCtx()
			eventsReceiver.Close(closeCtx)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		eventsReceiver, err := s.eventsReceiverFactory.NewReceiver(projectID)
		if err != nil { // It's fatal if we can't get an event receiver
			select {
			case s.errCh <- err:
			case <-ctx.Done():
			}
			return
		}

		// This is the main loop for receiving this project's events
		for {
			// Get the next event for this project
			asyncEvent, err := eventsReceiver.Receive(ctx)
			if err != nil {
				continue outerLoop // Try again with a new receiver
			}

			event, err := s.apiClient.Events().Get(ctx, asyncEvent.EventID)
			if err != nil {
				// TODO: We should check what went wrong
				log.Println(err)
				asyncEvent.Ack()
				continue // Next event
			}

			// If the worker's phase isn't PENDING, then there's nothing to do
			if event.Status.WorkerStatus.Phase != sdk.WorkerPhasePending {
				asyncEvent.Ack()
				continue // Next event
			}

			// TODO: We should still check k8s for the existence of the pod before
			// proceeding, because with at least once event delivery semantics, there
			// is always the possibility that we already scheduled this pod, but the
			// worker's status remains PENDING only because the observer is down.

			// TODO: Wait for project capacity

			// Wait for system capacity
			select {
			case <-s.availabilityCh:
			case <-ctx.Done():
				// We don't ack the event here because it hasn't been scheduled yet
				continue outerLoop // This will do cleanup before returning
			}

			// Now start the worker...

			if err := s.createWorkspacePVC(ctx, event); err != nil {
				// TODO: Update the event in the database to reflect the error
				log.Printf(
					"error creating workspace for event %q worker: %s",
					asyncEvent.EventID,
					err,
				)
				asyncEvent.Ack()
				continue // Next event
			}

			if err := s.createWorkerPod(ctx, event); err != nil {
				// TODO: Update the event in the database to reflect the error
				log.Printf(
					"error creating pod for event %q worker: %s",
					asyncEvent.EventID,
					err,
				)
				asyncEvent.Ack()
				continue // Next event
			}

			asyncEvent.Ack()
		}

	}

}
