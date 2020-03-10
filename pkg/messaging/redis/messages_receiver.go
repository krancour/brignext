package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/krancour/brignext/pkg/messaging"
)

// defaultReceivePendingMessages receives message IDs from the global pending
// list and transplants them to a consumer-specific active list. It retrieves
// the corresponding message from the global messages hash and dispatches it
// over a channel for processing.
func (c *consumer) defaultReceivePendingMessages(ctx context.Context) {
	defer c.wg.Done()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
outer:
	for {
		// Don't try getting work off the pending list until we know we have a
		// handler that's ready to do work, otherwise we could end up claiming
		// work that we're not ready to do.
		select {
		case <-c.handlerReadyCh:
		case <-ctx.Done():
			return
		}
		// TODO: Count failures
		for {
			res := c.redisClient.RPopLPush(
				c.pendingListName,
				c.activeListName,
			)
			err := res.Err()
			if err == redis.Nil {
				select {
				// This delay before trying again when we've just found nothing stops us
				// from taxing the CPU, network, or database when the pending list is
				// empty.
				case <-time.After(*c.options.ReceiverNoResultPauseInterval):
					continue
				case <-ctx.Done():
					return
				}
			}
			if err != nil {
				c.abort(ctx, err)
				return
			}
			messageID := res.Val()
			messageJSON, err :=
				c.redisClient.HGet(c.messagesHashName, messageID).Bytes()
			if err != nil {
				// TODO: Distinguish between a real failure and a nil result.
				c.abort(ctx, err)
				return
			}
			message, err := messaging.NewMessageFromJSON(messageJSON)
			if err != nil {
				// TODO: Don't abort here. This is a poison message. Discard it or
				// maybe move it to a dead letter queue.
				c.abort(ctx, err)
				return
			}
			select {
			case c.messageCh <- message:
				continue outer
			case <-ctx.Done():
				return
			}
		}
	}
}
