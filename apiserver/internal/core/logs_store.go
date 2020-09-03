package core

import (
	"context"
)

type LogsStore interface {
	StreamLogs(
		ctx context.Context,
		event Event,
		selector LogsSelector,
		opts LogStreamOptions,
	) (<-chan LogEntry, error)
}
