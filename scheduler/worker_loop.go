package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/krancour/brignext/v2/scheduler/internal/queue"
	"github.com/krancour/brignext/v2/sdk"
)

func (s *scheduler) runWorkerLoop(ctx context.Context, projectID string) {

	var workersReader queue.Reader

outerLoop:
	for {

		if workersReader != nil {
			func() {
				closeCtx, cancelCloseCtx :=
					context.WithTimeout(context.Background(), 5*time.Second)
				defer cancelCloseCtx()
				workersReader.Close(closeCtx)
			}()
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		var err error
		workersReader, err = s.queueReaderFactory.NewQueueReader(
			fmt.Sprintf("workers.%s", projectID),
		)
		if err != nil { // It's fatal if we can't get a queue reader
			select {
			case s.errCh <- err:
			case <-ctx.Done():
			}
			return
		}

		// This is the main loop for receiving this Project's Events' Workers
		for {
			// Get the next Worker
			msg, err := workersReader.Read(ctx)
			if err != nil {
				continue outerLoop // Try again with a new reader
			}

			event, err := s.apiClient.Events().Get(ctx, msg.Message)
			if err != nil {
				// TODO: We should check what went wrong
				log.Println(err)
				msg.Ack() // nolint: errcheck
				continue  // Next event
			}

			// If the worker's phase isn't PENDING, then there's nothing to do
			if event.Worker.Status.Phase != sdk.WorkerPhasePending {
				msg.Ack() // nolint: errcheck
				continue  // Next event
			}

			// TODO: We should still check k8s for the existence of the pod before
			// proceeding, because with at least once event delivery semantics, there
			// is always the possibility that we already scheduled this pod, but the
			// worker's status remains PENDING only because the observer is down...
			// But the API should do that.

			// TODO: Wait for Project capacity

			// Wait for system capacity
			select {
			case <-s.availabilityCh:
			case <-ctx.Done():
				// We don't ack the event here because it hasn't been scheduled yet
				continue outerLoop // This will do cleanup before returning
			}

			// Now use the API to start the Worker...

			if err := s.apiClient.Events().StartWorker(ctx, event.ID); err != nil {
				log.Printf(
					"error starting worker for event %q: %s",
					msg.Message,
					err,
				)
			}

			msg.Ack() // nolint: errcheck
		}

	}

}
