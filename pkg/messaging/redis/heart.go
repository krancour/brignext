package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis"
)

// defaultRunHeart emits "heartbeats" at regular intervals as proof of life.
func (c *consumer) defaultRunHeart(ctx context.Context) {
	defer c.wg.Done()
	ticker := time.NewTicker(*c.options.HeartbeatInterval)
	defer ticker.Stop()
	for {
		if ok := c.manageRetries(
			ctx,
			"send heartbeat",
			*c.options.HeartbeatMaxAttempts,
			30*time.Second, // TODO: Don't hardcode this
			c.heartbeat,
		); !ok {
			return
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

// heartbeat adds/updates a member in a sorted set, scored by the current time.
// When this consumer inevitably dies, replacement consumers' cleaning processes
// will easily be able to recognize it as dead.
func (c *consumer) heartbeat() error {
	return c.redisClient.ZAdd(
		c.consumersSetName,
		redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: c.activeListName,
		},
	).Err()
}
