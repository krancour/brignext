package redis

import (
	"context"
	"time"
)

const cleaningInterval = time.Minute

// defaultRunCleaner continuously monitors the heartbeats of all known consumers
// for proof of life. When a known consumer is found to have died, incomplete
// work assigned to the dead consumer will be transplanted back to the global
// pending message list.
func (c *consumer) defaultRunCleaner(ctx context.Context) {
	defer c.wg.Done()
	ticker := time.NewTicker(cleaningInterval)
	defer ticker.Stop()
	for {
		if err := c.redisClient.EvalSha(
			c.cleanerScriptSHA,
			[]string{c.consumersSetName, c.pendingListName},
			// TODO: This is any consumer that has missed even one heartbeat.
			// This may be may too harsh. Make this configurable?
			time.Now().Add(-2*heartbeatInterval).Unix(),
		).Err(); err != nil {
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
