package events

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/core"
)

type LogsStore interface {
	StreamLogs(
		ctx context.Context,
		event core.Event,
		selector core.LogsSelector,
		opts core.LogStreamOptions,
	) (<-chan core.LogEntry, error)
}
