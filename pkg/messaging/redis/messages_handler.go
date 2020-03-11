package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
)

// defaultHandleMessages receives messages over a channel and delegates message
// handling to a user-defined handler function. Errors returned by the
// user-defined function are considered non-fatal and are logged. Redis-related
// failures-- e.g. a failure removing the handled message or its ID from
// relevant data structures-- are fatal and will cause this function to return.
func (c *consumer) defaultHandleMessages(ctx context.Context) {
	defer c.wg.Done()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for {
		// Let the receiver know we're ready to work. This ensures that the receiver
		// isn't too eager to claim work that we're not ready to handle.
		select {
		case c.handlerReadyCh <- struct{}{}:
		case <-ctx.Done():
			return
		}
		select {
		case message := <-c.messageCh:
			if err := c.handler(ctx, message); err != nil {
				if err == ctx.Err() {
					// The error is simply that the user-defined handler function was
					// preempted and that function had the good sense to return ctx.Err().
					// Just return and let a replacement consumer's cleaner process deal
					// with re-queuing the work.
					return
				}
				// If we get to here, this is a legitimate failure handling the message.
				// This isn't the consumer's fault. Simply log this.
				log.Println(
					errors.Wrapf(
						err,
						"queue %q consumer %q encountered an error handling message %q",
						c.baseQueueName,
						c.id,
						message.ID(),
					),
				)
			} else {
				// There was no error returned from the user-defined handler function,
				// but it's POSSIBLE the context was canceled and the that function was
				// preempted, but it didn't return ctx.Err(). We'll check if the context
				// has been canceled, and if it has, we will assume this worst case
				// scenario and just return and let a replacement consumer's cleaner
				// process deal with re-queuing the work.
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
			// Error or no error, if we got to here, we know we're really done with
			// this message for good.
			if ok := c.manageRetries(
				ctx,
				fmt.Sprintf("delete message %q", message.ID()),
				*c.options.ReceiverMaxAttempts, // TODO: This isn't the right option
				30*time.Second,                 // TODO: Don't hardcode this,
				func() error {
					return c.deleteMessage(message.ID())
				},
			); !ok {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *consumer) deleteMessage(messageID string) error {
	pipeline := c.redisClient.TxPipeline()
	pipeline.LRem(c.activeListName, -1, messageID)
	pipeline.HDel(c.messagesHashName, messageID)
	_, err := pipeline.Exec()
	return err
}
