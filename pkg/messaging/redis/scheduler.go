package redis

import (
	"context"
	"time"
)

// TODO: Make this configurable
var schedulerInterval = 5 * time.Second

// defaultRunScheduler checks at regular intervals for messages in the global
// scheduled set that can be moved to the global pending list.
func (c *consumer) defaultRunScheduler(ctx context.Context) {
	defer c.wg.Done()
	ticker := time.NewTicker(schedulerInterval)
	defer ticker.Stop()
	for {
		if err := c.redisClient.EvalSha(
			c.schedulerScriptSHA,
			[]string{c.scheduledSetName, c.pendingListName},
			float64(time.Now().Unix()),
			50, // TODO: Make this configurable?
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
