package redis

import (
	"fmt"
)

func prefixedKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", prefix, key)
}

func pendingListKey(prefix, queueName string) string {
	return prefixedKey(
		prefix,
		fmt.Sprintf("%s:pending", queueName),
	)
}

func messagesHashKey(prefix, queueName string) string {
	return prefixedKey(
		prefix,
		fmt.Sprintf("%s:messages", queueName),
	)
}

func scheduledSetKey(prefix, queueName string) string {
	return prefixedKey(
		prefix,
		fmt.Sprintf("%s:scheduled", queueName),
	)
}

func consumersSetKey(prefix, queueName string) string {
	return prefixedKey(
		prefix,
		fmt.Sprintf("%s:consumers", queueName),
	)
}

func activeListKey(prefix, queueName, consumerID string) string {
	return prefixedKey(
		prefix,
		fmt.Sprintf("%s:%s", queueName, consumerID),
	)
}
