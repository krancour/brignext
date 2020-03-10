package redis

import (
	"context"
	"fmt"
)

func prefixedName(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", prefix, key)
}

func pendingListName(prefix, baseName string) string {
	return prefixedName(
		prefix,
		fmt.Sprintf("%s:pending", baseName),
	)
}

func messagesHashName(prefix, baseName string) string {
	return prefixedName(
		prefix,
		fmt.Sprintf("%s:messages", baseName),
	)
}

func scheduledSetName(prefix, baseName string) string {
	return prefixedName(
		prefix,
		fmt.Sprintf("%s:scheduled", baseName),
	)
}

func consumersSetName(prefix, baseName string) string {
	return prefixedName(
		prefix,
		fmt.Sprintf("%s:consumers", baseName),
	)
}

func activeListName(prefix, baseName, consumerID string) string {
	return prefixedName(
		prefix,
		fmt.Sprintf("%s:%s:active", baseName, consumerID),
	)
}

func heartbeatKey(prefix, baseName, consumerID string) string {
	return prefixedName(
		prefix,
		fmt.Sprintf("%s:%s:heartbeat", baseName, consumerID),
	)
}

func (c *consumer) abort(ctx context.Context, err error) {
	select {
	case c.errCh <- err:
	case <-ctx.Done():
	}
}