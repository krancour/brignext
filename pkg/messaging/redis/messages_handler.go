package redis

import (
	"context"
	"log"

	"github.com/krancour/brignext/pkg/messaging"
	"github.com/pkg/errors"
)

// defaultHandleMessages receives message bytes over a channel, decodes them,
// and delegates message handling to the consumer's handler function. Errors in
// handling are non-fatal and are logged. Redis-related failures-- e.g. a
// failure removing the handled message from the queue-- are fatal and will
// cause this function to return.
func (c *consumer) defaultHandleMessages(
	ctx context.Context,
	messageCh <-chan []byte,
	handler messaging.HandlerFn,
	errCh chan<- error,
) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for {
		select {
		case messageJSON := <-messageCh:
			message, err := c.getMessageFromJSON(
				messageJSON,
				c.activeQueueName,
			)
			if err != nil {
				select {
				case errCh <- err:
				case <-ctx.Done():
				}
				return // This error is fatal
			}
			if message == nil {
				// If the message is nil, it's because it was malformed. Just move on.
				continue
			}
			if err := handler(ctx, message); err != nil {
				// If we get to here, we have a legitimate failure handling the error.
				// This isn't the consumer's fault. Simply log this.
				log.Println(
					errors.Wrapf(
						err,
						"consumer %q encountered an error handling message %q",
						c.id,
						message.ID(),
					),
				)
			}
			// Regardless of success or failure, we're done with this message. Remove
			// it from the active message queue. This is the reason we held on to the
			// original message bytes all the way through the process.
			if err := c.redisClient.LRem( // Remove by value
				c.activeQueueName,
				-1,
				messageJSON,
			).Err(); err != nil {
				select {
				case errCh <- errors.Wrapf(
					err,
					"error removing message %q from queue %q",
					message.ID(),
					c.activeQueueName,
				):
				case <-ctx.Done():
				}
				return // This error is fatal
			}
		case <-ctx.Done():
			return
		}
	}
}
