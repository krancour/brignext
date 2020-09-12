package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/brigadecore/brigade/v2/scheduler/internal/queue"
	"github.com/brigadecore/brigade/v2/sdk/core"
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

			event, err := s.coreClient.Events().Get(ctx, msg.Message)
			if err != nil {
				// TODO: We should check what went wrong
				log.Println(err)
				msg.Ack() // nolint: errcheck
				continue  // Next Worker
			}

			// If the Worker's phase isn't PENDING, then there's nothing to do
			if event.Worker.Status.Phase != core.WorkerPhasePending {
				msg.Ack() // nolint: errcheck
				continue  // Next Worker
			}

			// TODO: Wait for Project capacity

			// Wait for system capacity
			select {
			case <-s.workerAvailabilityCh:
			case <-ctx.Done():
				// We don't ack the event here because it hasn't been scheduled yet
				continue outerLoop // This will do cleanup before returning
			}

			// Now use the API to start the Worker...

			if err :=
				s.coreClient.Events().Workers().Start(ctx, event.ID); err != nil {
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
