package redis

import (
	"context"
	"log"
	"time"

	"github.com/pkg/errors"
)

// defaultWatchDeferredMessages receives deferred message bytes over a channel,
// decodes them, and blocks until the specified time constraints on handling
// have been met. At that time, the message is moved to the pending queue. All
// errors are fatal and will cause this function to return.
func (c *consumer) defaultWatchDeferredMessages(
	ctx context.Context,
	messageCh <-chan []byte,
	errCh chan<- error,
) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for {
		select {
		case messageJSON := <-messageCh:
			message, err := c.getMessageFromJSON(
				messageJSON,
				c.watchedQueueName,
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
			handleTime := message.HandleTime()
			if handleTime == nil {
				// This message really shouldn't have been on this queue. Throw it
				// away and log it.
				err := c.redisClient.LRem(
					c.watchedQueueName,
					-1,
					messageJSON,
				).Err()
				if err != nil {
					select {
					case errCh <- errors.Wrapf(
						err,
						"error removing message %q with no handle time from queue %q",
						message.ID(),
						c.watchedQueueName,
					):
					case <-ctx.Done():
					}
					return // This error is fatal
				}
				log.Printf(
					"deferred message %q had no handle time and was removed from "+
						"queue %q",
					message.ID(),
					c.watchedQueueName,
				)
				continue // Move on
			}
			// Note if the duration passed to the timer is 0 or negative, it will go
			// off immediately
			timer := time.NewTimer(time.Until(*handleTime))
			defer timer.Stop()
			select {
			case <-timer.C:
				// Move the message to the pending queue
				pipeline := c.redisClient.TxPipeline()
				pipeline.LPush(c.pendingQueueName, messageJSON)
				pipeline.LRem(c.watchedQueueName, -1, messageJSON)
				_, err := pipeline.Exec()
				if err != nil {
					select {
					case errCh <- errors.Wrapf(
						err,
						"error moving deferred message %q to queue %q",
						message.ID(),
						c.pendingQueueName,
					):
					case <-ctx.Done():
					}
					return // This error is fatal
				}
			case <-ctx.Done():
			}
		case <-ctx.Done():
			return
		}
	}
}
