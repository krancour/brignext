package redis

import (
	"fmt"
	"log"

	"github.com/krancour/brignext/pkg/messaging"
	"github.com/pkg/errors"
)

func (c *consumer) getMessageFromJSON(
	messageJSON []byte,
	queueName string,
) (messaging.Message, error) {
	message, err := messaging.NewMessageFromJSON(messageJSON)
	if err != nil {
		// If the JSON is invalid, remove the message from the queue, log this and
		// move on. No other worker is going to be able to process this-- there's
		// nothing we can do and there's no sense treating this as a fatal
		// condition.
		err := c.redisClient.LRem(queueName, -1, messageJSON).Err()
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"error removing malformed message from queue %q",
				queueName,
			)
		}
		log.Println(
			errors.Wrapf(
				err,
				"error decoding malformed message from queue %q",
				queueName,
			),
		)
		return nil, nil
	}
	return message, nil
}

func prefixedName(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", prefix, key)
}

func pendingQueueName(prefix, baseName string) string {
	return prefixedName(
		prefix,
		fmt.Sprintf("%s:pending", baseName),
	)
}

func deferredQueueName(prefix, baseName string) string {
	return prefixedName(
		prefix,
		fmt.Sprintf("%s:deferred", baseName),
	)
}

func consumersSetName(prefix, baseName string) string {
	return prefixedName(
		prefix,
		fmt.Sprintf("%s:consumers", baseName),
	)
}

func activeQueueName(prefix, baseName, consumerID string) string {
	return prefixedName(
		prefix,
		fmt.Sprintf("%s:%s:active", baseName, consumerID),
	)
}

func watchedQueueName(prefix, baseName, consumerID string) string {
	return prefixedName(
		prefix,
		fmt.Sprintf("%s:%s:watched", baseName, consumerID),
	)
}

func heartbeatKey(prefix, baseName, consumerID string) string {
	return prefixedName(
		prefix,
		fmt.Sprintf("%s:%s:heartbeat", baseName, consumerID),
	)
}
