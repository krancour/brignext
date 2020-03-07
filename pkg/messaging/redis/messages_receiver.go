package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// defaultReceiveMessages receives messages from a source queue and dispatches
// them to a to both a destination queue and a message channel.
func (c *consumer) defaultReceiveMessages(
	ctx context.Context,
	sourceQueueName string,
	destinationQueueName string,
	messageCh chan<- []byte,
	errCh chan<- error,
) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for {
		messageJSON, err := c.redisClient.BRPopLPush(
			sourceQueueName,
			destinationQueueName,
			// Don't try indefinitely, or else we'll never have the opportunity to
			// exit if/when context is canceled
			time.Second*5,
		).Bytes()
		if err == redis.Nil {
			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
		if err != nil {
			select {
			case errCh <- errors.Wrapf(
				err,
				"error receiving message from queue %q",
				sourceQueueName,
			):
			case <-ctx.Done():
			}
			return // This error is fatal
		}
		select {
		// It may seem odd that we're putting []byte on the channel instead of a
		// messaging.Message, but there's a reason... Later, when the message is
		// handled, whether successfully or unsuccessfully, the message needs to be
		// removed from the queue BY VALUE-- i.e. we need to know the original
		// message bytes in order to remove it.
		case messageCh <- messageJSON:
		case <-ctx.Done():
			return
		}
	}
}
