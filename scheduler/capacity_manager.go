package main

import (
	"context"
	"time"
)

// manageCapacity periodically checks how many worker pods are currently running
// and sends a signal on an availability channel when there is available
// capacity.
func (s *scheduler) manageCapacity(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		s.syncMu.Lock()
		runningWorkerPods := len(s.workerPodsSet)
		// Give up this lock before we potentially block waiting on someone who's
		// ready for the capacity we might be allocating.
		s.syncMu.Unlock()
		// TODO: Make this configurable
		if runningWorkerPods < 2 {
			select {
			case s.availabilityCh <- struct{}{}:
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
