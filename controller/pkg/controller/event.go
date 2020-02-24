package controller

import (
	"context"
	"log"
	"time"

	"github.com/deis/async"
	"github.com/krancour/brignext"
	"github.com/pkg/errors"
)

func (c *controller) processEvent(
	ctx context.Context,
	task async.Task,
) ([]async.Task, error) {
	// TODO: Remove this output
	log.Printf(
		"received process event task for event %q",
		task.GetArgs()["eventID"],
	)

	eventID, ok := task.GetArgs()["eventID"]
	if !ok {
		return nil, errors.Errorf(
			"processEvent task %q did not include an event ID argument",
			task.GetID(),
		)
	}

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

	if event.Status == brignext.EventStatusAccepted {
		return c.processAcceptedEvent(ctx, event)
	}

	return nil, nil
}

func (c *controller) processAcceptedEvent(
	ctx context.Context,
	event brignext.Event,
) ([]async.Task, error) {
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

	workers := project.GetWorkers(event.Provider, event.Type)
	var status brignext.EventStatus
	if len(workers) == 0 {
		status = brignext.EventStatusMoot
	} else {
		status = brignext.EventStatusProcessing
	}

	if err := c.apiClient.UpdateEventWorkers(ctx, event.ID, workers); err != nil {
		return nil, errors.Wrapf(
			err,
			"error updating event %q workers",
			event.ID,
		)
	}

	if err := c.apiClient.UpdateEventStatus(ctx, event.ID, status); err != nil {
		return nil, errors.Wrapf(
			err,
			"error updating event %q status",
			event.ID,
		)
	}

	tasks := make([]async.Task, len(workers))
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

	return tasks, nil
}
