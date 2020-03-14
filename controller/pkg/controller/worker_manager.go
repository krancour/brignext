package controller

import (
	"context"
	"encoding/json"

	"github.com/krancour/brignext/pkg/messaging"
	"github.com/pkg/errors"
)

func (c *controller) handleProjectWorkerMessage(
	ctx context.Context,
	msg messaging.Message,
) error {
	workerContext := workerContext{}
	if err := json.Unmarshal(msg.Body(), &workerContext); err != nil {
		return errors.Wrap(
			err,
			"error unmarshaling message body into worker context",
		)
	}
	workerContext.doneCh = make(chan struct{})
	// Try to hand off worker execution
	select {
	case c.workerContextCh <- workerContext:
	case <-ctx.Done():
	}
	// Wait for worker execution to complete
	select {
	case <-workerContext.doneCh:
	case <-ctx.Done():
	}
	return nil
}

// TODO: Implement this
func (c *controller) defaultManageWorkers(ctx context.Context) {

	// TODO: This needs to receive work on the c.workerContextCh

}
