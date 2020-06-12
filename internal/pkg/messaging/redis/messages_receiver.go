package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
	"github.com/krancour/brignext/v2/internal/pkg/messaging"
	"github.com/krancour/brignext/v2/internal/pkg/retries"
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
			var messageID string
			if err := retries.ManageRetries(
				ctx,
				"deque a pending message",
				*c.options.RedisOperationMaxAttempts,
				*c.options.RedisOperationMaxBackoff,
				func() (bool, error) {
					var err error
					messageID, err = c.dequeueMessage()
					if err != nil {
						return true, err // Retry
					}
					return false, nil // No retry
				},
			); err != nil {
				select {
				case c.errCh <- err:
				case <-ctx.Done():
				}
				return
			}

			if messageID == "" {
				select {
				// This delay stops us from taxing the CPU, network, or database when
				// the pending list is empty.
				case <-time.After(*c.options.ReceiverNoResultPauseInterval):
					continue
				case <-ctx.Done():
					return
				}
			}

			var messageJSON []byte
			if err := retries.ManageRetries(
				ctx,
				fmt.Sprintf("retrieve message %q", messageID),
				*c.options.RedisOperationMaxAttempts,
				*c.options.RedisOperationMaxBackoff,
				func() (bool, error) {
					var err error
					messageJSON, err = c.getMessageJSON(messageID)
					if err != nil {
						return true, err // Retry
					}
					return false, nil // No retry
				},
			); err != nil {
				select {
				case c.errCh <- err:
				case <-ctx.Done():
				}
				return
			}

			if messageJSON == nil {
				// Somehow we couldn't find a message with the indicated ID. Log this
				// as a warning and move on.
				log.Printf(
					"ERROR: queue %q consumer %q could not locate message %q",
					c.queueName,
					c.id,
					messageID,
				)
				continue
			}

			message, err := messaging.NewMessageFromJSON(messageJSON)
			if err != nil {
				log.Printf(
					"ERROR: queue %q consumer %q failed to decode message %q: %s",
					c.queueName,
					c.id,
					messageID,
					err,
				)
				continue
			}

			// Finally, deliver the message to a waiting handler
			select {
			case c.messageCh <- message:
				continue outer
			case <-ctx.Done():
				return
			}
		}
	}
}

func (c *consumer) dequeueMessage() (string, error) {
	messageID, err := c.redisClient.RPopLPush(
		c.pendingListKey,
		c.activeListKey,
	).Result()
	if err == redis.Nil {
		return "", nil
	}
	return messageID, err
}

func (c *consumer) getMessageJSON(messageID string) ([]byte, error) {
	messageJSON, err :=
		c.redisClient.HGet(c.messagesHashKey, messageID).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return messageJSON, err
}
