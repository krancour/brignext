package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

const cleaningInterval = time.Second * 30

// defaultClean continuously monitors the heartbeats of all known consumers for
// proof of life. When a known consumer is found to have died, incomplete work
// assigned to the dead consumer will be transplanted to the applicable global
// queues. defaultClean always returns a non-nil error.
func (c *consumer) defaultClean(ctx context.Context) error {
	ticker := time.NewTicker(cleaningInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			consumerIDs, err := c.redisClient.SMembers(c.consumersSetName).Result()
			if err == redis.Nil {
				continue
			}
			if err != nil {
				return errors.Wrapf(
					err,
					"error retrieving queue %q consumers",
					c.baseQueueName,
				) // This is fatal
			}
			for _, consumerID := range consumerIDs {
				err := c.redisClient.Get(
					heartbeatKey(c.options.RedisPrefix, c.baseQueueName, consumerID),
				).Err()
				if err == nil {
					continue
				}
				if err != redis.Nil {
					return errors.Wrapf(
						err,
						"error checking health of queue %q consumer %q",
						c.baseQueueName,
						consumerID,
					) // This is fatal
				}
				// If we get to here, we have a dead consumer on our hands
				if err := c.cleanQueue(
					ctx,
					consumerID,
					activeQueueName(c.options.RedisPrefix, c.baseQueueName, consumerID),
					c.pendingQueueName,
				); err != nil {
					return errors.Wrapf(
						err,
						"error cleaning out active queue of dead consumer %q",
						consumerID,
					) // This is fatal
				}
				if err := c.cleanQueue(
					ctx,
					consumerID,
					watchedQueueName(c.options.RedisPrefix, c.baseQueueName, consumerID),
					c.deferredQueueName,
				); err != nil {
					return errors.Wrapf(
						err,
						"error cleaning out watched queue of dead consumer %q",
						consumerID,
					) // This is fatal
				}
				err = c.redisClient.SRem(c.consumersSetName, consumerID).Err()
				if err != nil && err != redis.Nil {
					return errors.Wrapf(
						err,
						"error removing dead consumer %q from queue %q consumers set",
						consumerID,
						c.baseQueueName,
					) // This is fatal
				}
			}
		case <-ctx.Done():
			return errors.Wrapf(
				ctx.Err(),
				"queue %q consumer %q cleaner shutting down",
				c.baseQueueName,
				c.id,
			)
		}
	}
}

// defaultCleanQueue cleans a given source queue by moving any/all messages
// contained within to a destination queue. This is useful, for instance, in
// moving incomplete work assigned to a dead consumer back to the applicable
// global queues.
func (c *consumer) defaultCleanQueue(
	ctx context.Context,
	consumerID string,
	sourceQueueName string,
	destinationQueueName string,
) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		// TODO: Is this the right operation? We want to make sure reclaimed work
		// goes at the head of the destination queue.
		err := c.redisClient.RPopLPush(sourceQueueName, destinationQueueName).Err()
		if err == redis.Nil {
			return nil
		}
		if err != nil {
			return errors.Wrapf(
				err,
				"error cleaning up after dead consumer %q queue %q",
				consumerID,
				sourceQueueName,
			)
		}
	}
}
