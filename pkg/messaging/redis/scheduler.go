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
	for {
		if ok := c.manageRetries(
			ctx,
			"schedule messages",
			c.schedule,
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

func (c *consumer) schedule() error {
	return c.redisClient.EvalSha(
		c.schedulerScriptSHA,
		[]string{c.scheduledSetKey, c.pendingListKey},
		float64(time.Now().Unix()),
		50, // Max number of messages to transplant in one shot
	).Err()
}
