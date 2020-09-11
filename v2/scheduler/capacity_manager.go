package main

import (
	"context"
	"time"
)

// manageWorkerCapacity periodically checks how many Worker pods are currently
// running and sends a signal on an availability channel when there is available
// capacity.
func (s *scheduler) manageWorkerCapacity(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		s.syncMu.Lock()
		runningWorkerPods := len(s.workerPodsSet)
		// Give up this lock before we potentially block someone who's otherwise
		// ready for the capacity we might be allocating.
		s.syncMu.Unlock()
		// TODO: Make this configurable
		if runningWorkerPods < 2 {
			select {
			case s.workerAvailabilityCh <- struct{}{}:
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

// manageJobCapacity periodically checks how many Job pods are currently running
// and sends a signal on an availability channel when there is available
// capacity.
func (s *scheduler) manageJobCapacity(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		s.syncMu.Lock()
		runningJobPods := len(s.jobPodsSet)
		// Give up this lock before we potentially block someone who's otherwise
		// ready for the capacity we might be allocating.
		s.syncMu.Unlock()
		// TODO: Make this configurable
		if runningJobPods < 2 {
			select {
			case s.jobAvailabilityCh <- struct{}{}:
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
