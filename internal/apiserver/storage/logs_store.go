package storage

import (
	"context"

	"github.com/krancour/brignext/v2"
)

type LogsStore interface {
	GetLogs(
		ctx context.Context,
		eventID string,
		opts brignext.LogOptions,
	) (brignext.LogEntryList, error)
	StreamLogs(
		ctx context.Context,
		eventID string,
		opts brignext.LogOptions,
	) (<-chan brignext.LogEntry, error)

	CheckHealth(context.Context) error
}
