package main

import (
	"context"
	"log"
	"time"

	redisMessaging "github.com/krancour/brignext/v2/internal/common/messaging/redis"
)

func (c *controller) defaultManageProjectWorkerQueueConsumers(
	ctx context.Context,
) {
	// Maintain a map of functions for canceling the contexts of queue consumers
	// for each known project.
	consumerContextCancelFns := map[string]func(){}

	// TODO: Is it ok that this is hardcoded?
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		projects, err := c.apiClient.GetProjects(ctx)
		if err != nil {
			select {
			case c.errCh <- err:
			case <-ctx.Done():
			}
			return
		}

		// Build a set of current projects. This makes it a little faster and easier
		// to search for projects later in this algorithm.
		currentProjects := map[string]struct{}{}
		for _, project := range projects {
			currentProjects[project.ID] = struct{}{}
		}

		// Reconcile differences between projects we knew about already and the
		// current set of projects...

		// 1. Stop queue consumers for projects that have been deleted
		for projectID, cancelFn := range consumerContextCancelFns {
			if _, stillExists := currentProjects[projectID]; !stillExists {
				log.Printf("DEBUG: stopping queue consumer for project %q", projectID)
				cancelFn()
				// Surprisingly, Go lets us delete from a map we are currently iterating
				// over. How convenient.
				delete(consumerContextCancelFns, projectID)
			}
		}

		// 2. Start queue consumers for projects that have been added
		var receiverCount uint8 = 1
		var handlerCount uint8 = 2 // TODO: Do not hardcode this
		for projectID := range currentProjects {
			if _, known := consumerContextCancelFns[projectID]; !known {
				log.Printf("DEBUG: starting queue consumer for project %q", projectID)
				consumer, err := redisMessaging.NewConsumer(
					c.redisClient,
					projectID,
					&redisMessaging.ConsumerOptions{
						LoneConsumer:             true,
						ConcurrentReceiversCount: &receiverCount,
						ConcurrentHandlersCount:  &handlerCount,
					},
					c.handleProjectWorkerMessage,
				)
				if err != nil {
					select {
					case c.errCh <- err:
					case <-ctx.Done():
					}
					return
				}
				consumerContext, consumerContextCancelFn := context.WithCancel(ctx)
				go func() {
					if err := consumer.Run(consumerContext); err != nil {
						select {
						case c.errCh <- err:
						case <-ctx.Done():
						}
					}
				}()
				consumerContextCancelFns[projectID] = consumerContextCancelFn
			}
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}

}
