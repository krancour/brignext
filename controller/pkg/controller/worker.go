package controller

import (
	"context"
	"log"

	"github.com/deis/async"
)

// TODO: Implement this
func (c *controller) executeWorker(
	ctx context.Context,
	task async.Task,
) ([]async.Task, error) {
	// TODO: Remove this output
	log.Printf(
		"received execute worker task for worker %q of event %q",
		task.GetArgs()["worker"],
		task.GetArgs()["eventID"],
	)
	return nil, nil
}
