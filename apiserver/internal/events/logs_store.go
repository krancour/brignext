package events

import (
	"context"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
)

type LogsStore interface {
	StreamLogs(
		ctx context.Context,
		event brignext.Event,
		selector brignext.LogsSelector,
		opts brignext.LogStreamOptions,
	) (<-chan brignext.LogEntry, error)
}
