package redis

import (
	"context"
	"time"
)

// defaultRunCleaner continuously monitors the heartbeats of all known consumers
// for proof of life. When a known consumer is found to have died, incomplete
// work assigned to the dead consumer will be transplanted back to the global
// pending message list.
func (c *consumer) defaultRunCleaner(ctx context.Context) {
	defer c.wg.Done()
	ticker := time.NewTicker(*c.options.CleanerInterval)
	defer ticker.Stop()
	for {
		if ok := c.manageRetries(
			ctx,
			"clean up after dead consumers",
			c.clean,
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

func (c *consumer) clean() error {
	// TODO: We should log how many messages were reclaimed
	return c.redisClient.EvalSha(
		c.cleanerScriptSHA,
		[]string{c.consumersSetKey, c.pendingListKey},
		time.Now().Add(-c.deadConsumerThreshold).Unix(),
	).Err()
}
