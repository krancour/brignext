package redis

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// defaultRunHeart emits heartbeats at regular intervals as proof of life to
// prevent other consumers' cleaners from reclaiming work currently assigned to
// this consumer. defaultRunHeart always returns a non-nil error.
func (c *consumer) defaultRunHeart(ctx context.Context) error {
	ticker := time.NewTicker(cleaningInterval)
	defer ticker.Stop()
	for {
		if err := c.heartbeat(); err != nil {
			return errors.Wrapf(
				err,
				"error sending heartbeat for queue %q consumer %q",
				c.baseQueueName,
				c.id,
			) // This is fatal
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return errors.Wrapf(
				ctx.Err(),
				"queue %q consumer %q heartbeat stopping",
				c.baseQueueName,
				c.id,
			)
		}
	}
}

// defaultHeartbeat emits a single heartbeat as proof of life to prevent other
// consumers' cleaners from reclaiming work currently assigned to this consumer.
func (c *consumer) defaultHeartbeat() error {
	const aliveIndicator = "alive"
	if err := c.redisClient.Set(
		c.heartbeatKey,
		aliveIndicator,
		cleaningInterval*2,
	).Err(); err != nil {
		return errors.Wrapf(
			err,
			"error sending heartbeat for queue %q consumer %q",
			c.baseQueueName,
			c.id,
		)
	}
	return nil
}
