package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

const heartbeatInterval = 30 * time.Second

// defaultRunHeart emits "heartbeats" at regular intervals as proof of life.
func (c *consumer) defaultRunHeart(ctx context.Context) {
	defer c.wg.Done()
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	for {
		if err := c.heartbeat(); err != nil {
			c.abort(ctx, err)
			return
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

// defaultHeartbeat adds/updates a member in a sorted set, scored by the current
// time. When this consumer inevitably dies, replacement consumers' cleaning
// processes will easily be able to identify any messages that this consumer
// died while handling.
func (c *consumer) defaultHeartbeat() error {
	if err := c.redisClient.ZAdd(
		c.consumersSetName,
		redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: c.activeListName,
		},
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