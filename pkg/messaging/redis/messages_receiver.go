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
		for {
			res := c.redisClient.RPopLPush(
				c.pendingListName,
				c.activeListName,
			)
			err := res.Err()
			if err == redis.Nil {
				select {
				// This delay before trying again when we've just found nothing stops us
				// from taxing the CPU or the network when the pending list is empty.
				// TODO: Make this configurable?
				case <-time.After(5 * time.Second):
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
				c.abort(ctx, err)
				return
			}
			message, err := messaging.NewMessageFromJSON(messageJSON)
			if err != nil {
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
