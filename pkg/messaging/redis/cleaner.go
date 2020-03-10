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
	var failureCount uint8
	for {
		// TODO: Loop if we didn't move all applicable messages.
		if err := c.redisClient.EvalSha(
			c.cleanerScriptSHA,
			[]string{c.consumersSetName, c.pendingListName},
			time.Now().Add(-*c.options.CleanerDeadConsumerThreshold).Unix(),
			// TODO: The script doesn't actually do anything with this next arg yet.
			50, // Max number of messages to transplant in one shot
		).Err(); err != nil {
			failureCount++
			if failureCount > *c.options.CleanerMaxFailures {
				c.abort(ctx, err)
				return
			}
		} else {
			failureCount = 0
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}
