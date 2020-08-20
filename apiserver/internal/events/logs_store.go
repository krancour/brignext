package events

import (
	"context"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
)

type LogsStore interface {
	GetLogs(
		ctx context.Context,
		event brignext.Event,
		opts brignext.LogOptions,
	) (brignext.LogEntryList, error)
	StreamLogs(
		ctx context.Context,
		event brignext.Event,
		opts brignext.LogOptions,
	) (<-chan brignext.LogEntry, error)
}
