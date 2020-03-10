package redis

import (
	"context"
	"time"
)

// defaultRunScheduler checks at regular intervals for messages in the global
// scheduled set that can be moved to the global pending list.
func (c *consumer) defaultRunScheduler(ctx context.Context) {
	defer c.wg.Done()
	ticker := time.NewTicker(*c.options.SchedulerInterval)
	defer ticker.Stop()
	var failedAttempts uint8
	for {
		// TODO: Loop if we didn't move all applicable messages.
		if err := c.redisClient.EvalSha(
			c.schedulerScriptSHA,
			[]string{c.scheduledSetName, c.pendingListName},
			float64(time.Now().Unix()),
			50, // Max number of messages to transplant in one shot
		).Err(); err != nil {
			failedAttempts++
			if failedAttempts == *c.options.SchedulerMaxAttempts {
				c.abort(ctx, err)
				return
			}
		} else {
			failedAttempts = 0
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}
