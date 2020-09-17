package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/v2/scheduler/internal/queue"
)

func (s *scheduler) runJobLoop(ctx context.Context, projectID string) {

	var jobsReader queue.Reader

outerLoop:
	for {

		if jobsReader != nil {
			func() {
				closeCtx, cancelCloseCtx :=
					context.WithTimeout(context.Background(), 5*time.Second)
				defer cancelCloseCtx()
				jobsReader.Close(closeCtx)
			}()
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		var err error
		jobsReader, err = s.queueReaderFactory.NewQueueReader(
			fmt.Sprintf("jobs.%s", projectID),
		)
		if err != nil { // It's fatal if we can't get a queue reader
			select {
			case s.errCh <- err:
			case <-ctx.Done():
			}
			return
		}

		// This is the main loop for receiving this Project's Events' Workers' Jobs
		for {
			// Get the next Job
			msg, err := jobsReader.Read(ctx)
			if err != nil {
				continue outerLoop // Try again with a new reader
			}

			messageTokens := strings.Split(msg.Message, ":")
			if len(messageTokens) != 2 {
				log.Printf(
					"received invalid message on project %q job queue",
					projectID,
				)
				msg.Ack() // nolint: errcheck
				continue  // Next Job
			}
			eventID := messageTokens[0]
			jobName := messageTokens[1]

			event, err := s.coreClient.Events().Get(ctx, eventID)
			if err != nil {
				// TODO: We should check what went wrong
				log.Println(err)
				msg.Ack() // nolint: errcheck
				continue  // Next Job
			}

			job, ok := event.Worker.Jobs[jobName]
			if !ok {
				log.Printf(
					"no job %q exists for event %q",
					jobName,
					eventID,
				)
				msg.Ack() // nolint: errcheck
				continue  // Next Job
			}

			// If the Job's phase isn't PENDING, then there's nothing to do
			if job.Status.Phase != core.JobPhasePending {
				msg.Ack() // nolint: errcheck
				continue  // Next Job
			}

			// TODO: Wait for Project capacity

			// Wait for system capacity
			select {
			case <-s.jobAvailabilityCh:
			case <-ctx.Done():
				// We don't ack the event here because it hasn't been scheduled yet
				continue outerLoop // This will do cleanup before returning
			}

			// Now use the API to start the Job...

			if err := s.coreClient.Events().Workers().Jobs().Start(
				ctx,
				event.ID,
				jobName,
			); err != nil {
				log.Printf(
					"error starting event %q job %q: %s",
					event.ID,
					jobName,
					err,
				)
			}

			msg.Ack() // nolint: errcheck
		}

	}

}
