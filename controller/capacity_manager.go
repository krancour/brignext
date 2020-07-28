package main

import (
	"context"
	"log"
	"time"
)

// defaultManageCapacity periodically checks how many worker pods are currently
// running and sends a signal on an availability channel when there is available
// capacity.
func (c *controller) manageCapacity(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		c.podsLock.Lock()
		runningWorkerPods := len(c.workerPodsSet)
		// Give up this lock before we potentially block waiting on someone who's
		// ready for the capacity we might be allocating.
		c.podsLock.Unlock()
		// TODO: Make this configurable
		if runningWorkerPods < 2 {
			log.Println("found available capacity")
			select {
			case c.availabilityCh <- struct{}{}:
			case <-ctx.Done():
				return
			}
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}
