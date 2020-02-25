package controller

import (
	"context"
	"log"

	"github.com/deis/async"
)

// TODO: Implement this
// -----------------------------------------------------------------------------
// eventMonitor conducts a periodic review of the statuses of all workers
// belonging to the event it is monitoring. When all workers have entered a
// terminal state, the event's final state can be determined and then updated
// using the API.
func (c *controller) eventMonitor(
	ctx context.Context,
	task async.Task,
) ([]async.Task, error) {
	log.Printf(
		"INFO: received monitorEvent task for event %q",
		task.GetArgs()["eventID"],
	)
	return nil, nil
}
