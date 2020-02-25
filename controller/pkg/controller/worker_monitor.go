package controller

import (
	"context"
	"log"

	"github.com/deis/async"
)

// TODO: Implement this
// -----------------------------------------------------------------------------
// workerMonitor listens to the k8s pod corresponding to the worker it is
// monitoring. When the pod has entered a terminal state, the worker's final
// state can be determined and then updated using the API.
func (c *controller) workerMonitor(
	ctx context.Context,
	task async.Task,
) ([]async.Task, error) {
	log.Printf(
		"INFO: received monitorWorker task for worker %q of event %q",
		task.GetArgs()["worker"],
		task.GetArgs()["eventID"],
	)
	return nil, nil
}
