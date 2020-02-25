package controller

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/deis/async"
	"github.com/krancour/brignext"
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// workerExecute launches a pod corresponding to the specified worker then
// schedules a follow-up task to monitor that pod for completion.
func (c *controller) workerExecute(
	ctx context.Context,
	task async.Task,
) ([]async.Task, error) {
	eventID, ok := task.GetArgs()["eventID"]
	if !ok {
		return nil, errors.Errorf(
			"executeWorker task %q did not include an event ID argument",
			task.GetID(),
		)
	}

	workerName, ok := task.GetArgs()["worker"]
	if !ok {
		return nil, errors.Errorf(
			"executeWorker task %q did not include a worker name argument",
			task.GetID(),
		)
	}

	log.Printf(
		"INFO: received executeWorker task for worker %q of event %q",
		workerName,
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
			"error retrieving event %q for worker %q execution",
			eventID,
			workerName,
		)
	}

	worker, ok := event.Workers[workerName]
	if !ok {
		return nil, errors.Errorf(
			"executeWorker task %q failed because event %q did not have a worker "+
				"named %q",
			task.GetID(),
			eventID,
			workerName,
		)
	}

	// There's an unlikely, but non-zero possibility that this handler runs with
	// the worker status already, unexpectedly in a RUNNING state. This could only
	// happen if the handler has already run for this event at least once before
	// and succeeded in updating the worker's status in the database, but the
	// controller process exited unexpectedly before the task completed-- and
	// hence before the follow-up task to monitor the worker was added to the
	// async engine's work queue.
	//
	// So...
	//
	// If the status is already RUNNING, don't do any updates to the database.
	// Just return the follow-up tasks to monitor the worker.
	if worker.Status == brignext.WorkerStatusRunning {
		return []async.Task{
			// A task that will monitor the worker
			async.NewTask(
				"monitorWorker",
				map[string]string{
					"eventID": event.ID,
					"worker":  workerName,
				},
			),
		}, nil
	}

	// If the event status is anything other than PENDING, just log it and move
	// on.
	if worker.Status != brignext.WorkerStatusPending {
		log.Printf(
			"WARNING: worker %q of event %q status was unexpectedly %q when "+
				"initiating worker execution. Taking no action and moving on.",
			workerName,
			event.ID,
			worker.Status,
		)
		return nil, nil
	}

	// Get the worker pod up and running, if it isn't already
	if err := c.ensureWorkerPodRunning(ctx, event, workerName); err != nil {
		return nil, errors.Wrapf(
			err,
			"error ensuring worker pod running for worker %q of event %q",
			workerName,
			event.ID,
		)
	}

	// Update the worker status in the database
	if err := c.apiClient.UpdateEventWorkerStatus(
		ctx,
		eventID,
		workerName,
		brignext.WorkerStatusRunning,
	); err != nil {
		return nil, errors.Wrapf(
			err,
			"error updating status on worker %q of event %q",
			workerName,
			event.ID,
		)
	}

	// Schedule a task that will monitor the worker
	return []async.Task{
		// A task that will monitor the worker
		async.NewDelayedTask(
			"monitorWorker",
			map[string]string{
				"eventID": event.ID,
				"worker":  workerName,
			},
			5*time.Second,
		),
	}, nil
}

func (c *controller) ensureWorkerPodRunning(
	ctx context.Context,
	event brignext.Event,
	workerName string,
) error {
	podName := getWorkerPodName(event.ID, workerName)

	podsClient := c.kubeClient.CoreV1().Pods(event.Namespace)

	pod, err := podsClient.Get(podName, meta_v1.GetOptions{})
	if err != nil {
		return errors.Wrapf(
			err,
			"error checking for existing pod named %q in namespace %q",
			podName,
			event.Namespace,
		)
	}

	if pod != nil {
		log.Printf(
			"WARNING: pod named %q in namespace %q unexpectedly exists already",
			podName,
			event.Namespace,
		)
		return nil
	}

	return c.createWorkerPod(ctx, event, workerName)
}

// TODO: Implement this
// krancour: Finishing the implementation here is the magic that will expose
// whatever isn't right with the new domain model.
func (c *controller) createWorkerPod(
	ctx context.Context,
	event brignext.Event,
	workerName string,
) error {
	return nil
}

func getWorkerPodName(eventID, workerName string) string {
	return fmt.Sprintf("%s-%s", eventID, workerName)
}
