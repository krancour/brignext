package controller

import (
	"context"
	"log"

	"github.com/deis/async"
)

// TODO: Implement this
func (c *controller) processEvent(
	ctx context.Context,
	task async.Task,
) ([]async.Task, error) {
	log.Printf("received event ID %q", task.GetArgs()["eventID"])
	return nil, nil
}
