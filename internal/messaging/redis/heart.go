package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/krancour/brignext/v2/internal/retries"
)

// defaultRunHeart emits "heartbeats" at regular intervals as proof of life.
func (c *consumer) defaultRunHeart(ctx context.Context) {
	defer c.wg.Done()
	ticker := time.NewTicker(*c.options.HeartbeatInterval)
	defer ticker.Stop()
	for {
		if err := retries.ManageRetries(
			ctx,
			"send heartbeat",
			*c.options.RedisOperationMaxAttempts,
			*c.options.RedisOperationMaxBackoff,
			func() (bool, error) {
				if err := c.heartbeat(); err != nil {
					return true, err // Retry
				}
				return false, nil // No retry
			},
		); err != nil {
			select {
			case c.errCh <- err:
			case <-ctx.Done():
			}
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
		c.consumersSetKey,
		redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: c.activeListKey,
		},
	).Err()
}
