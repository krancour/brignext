package controller

import (
	"context"
	"log"
	"time"

	"github.com/deis/async"
	"github.com/krancour/brignext"
	"github.com/pkg/errors"
)

// eventProcess transitions an event from the ACCEPTED state to the PROCESSING
// state by determining, based on event details and project configuration, which
// workers need to be executed. These workers are added to the database, event
// status is updates, and asynchronous execution of the workers is scheduled.
func (c *controller) eventProcess(
	ctx context.Context,
	task async.Task,
) ([]async.Task, error) {
	eventID, ok := task.GetArgs()["eventID"]
	if !ok {
		return nil, errors.Errorf(
			"processEvent task %q did not include an event ID argument",
			task.GetID(),
		)
	}

	log.Printf(
		"INFO: received processEvent task for event %q",
		eventID,
	)

	// Find the event
	event, err := c.apiClient.GetEvent(ctx, eventID)
	if err != nil {
		if _, ok := err.(*brignext.ErrEventNotFound); ok {
			// The event wasn't found. The likely scenario is that it was deleted.
			// We're not going to treat this as an error. We're just going to move on.
			return nil, nil
		}
		return nil, errors.Wrapf(
			err,
			"error retrieving event %q for processing",
			eventID,
		)
	}

	// There's an unlikely, but non-zero possibility that this handler runs with
	// the event status already, unexpectedly in a PROCESSING state. This could
	// only happen if the handler has already run for this event at least once
	// before and succeeded in updating the workers and event status in the
	// database, but the controller process exited unexpectedly before the task
	// completed-- and hence before follow-up tasks to execute workers and monitor
	// the event were added to the async engine's work queue.
	//
	// So...
	//
	// If the status is already PROCESSING, don't do any updates to the database.
	// Just return the follow-up tasks to execute workers and monitor the event.
	if event.Status == brignext.EventStatusProcessing {
		tasks := make([]async.Task, len(event.Workers)+1)
		var i int
		for workerName := range event.Workers {
			tasks[i] = async.NewTask(
				"executeWorker",
				map[string]string{
					"eventID": event.ID,
					"worker":  workerName,
				},
			)
			i++
		}
		// A task that will monitor the event
		tasks[len(event.Workers)] = async.NewDelayedTask(
			"monitorEvent",
			map[string]string{
				"eventID": event.ID,
			},
			5*time.Second,
		)
		return tasks, nil
	}

	// If the event status is anything other than ACCEPTED, just log it and move
	// on.
	if event.Status != brignext.EventStatusAccepted {
		log.Printf(
			"WARNING: event %q status was unexpectedly %q when initiating event "+
				"processing. Taking no action and moving on.",
			event.ID,
			event.Status,
		)
		return nil, nil
	}

	project, err := c.apiClient.GetProject(ctx, event.ProjectID)
	if err != nil {
		if _, ok := err.(*brignext.ErrProjectNotFound); ok {
			// The project wasn't found. The likely scenario is that it was deleted.
			// We're not going to treat this as an error. We're just going to move on.
			return nil, nil
		}
		return nil, errors.Wrapf(
			err,
			"error retrieving project %q for event %q processing",
			event.ProjectID,
			event.ID,
		)
	}

	if err := c.apiClient.CreateEventSecrets(ctx, event.ID); err != nil {
		return nil, errors.Wrapf(
			err,
			"error creating event secrets for event %q processing",
			event.ID,
		)
	}

	// "Split" the event into many workers and update the event (workers and
	// status) using the API.
	workers := project.GetWorkers(event)
	if err := c.apiClient.UpdateEventWorkers(ctx, event.ID, workers); err != nil {
		return nil, errors.Wrapf(
			err,
			"error updating event %q workers",
			event.ID,
		)
	}
	var status brignext.EventStatus
	if len(workers) == 0 {
		status = brignext.EventStatusMoot
	} else {
		status = brignext.EventStatusProcessing
	}

	// Updte the event status in the database
	if err := c.apiClient.UpdateEventStatus(ctx, event.ID, status); err != nil {
		return nil, errors.Wrapf(
			err,
			"error updating event %q status",
			event.ID,
		)
	}

	// Return follow-up tasks for executing workers and monitoring the event
	// status.
	tasks := make([]async.Task, len(workers)+1)
	var i int
	for workerName := range workers {
		// TODO: Fix this
		// There's deliberately a short delay here to minimize the possibility of
		// the controller trying (and failing) to locate this worker before
		// the transaction that updated the event has become durable.
		tasks[i] = async.NewDelayedTask(
			"executeWorker",
			map[string]string{
				"eventID": event.ID,
				"worker":  workerName,
			},
			5*time.Second,
		)
		i++
	}
	// A task that will monitor the event
	tasks[len(workers)] = async.NewDelayedTask(
		"monitorEvent",
		map[string]string{
			"eventID": event.ID,
		},
		5*time.Second,
	)

	return tasks, nil
}
