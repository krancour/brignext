package events

import (
	"context"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
)

type LogsStore interface {
	GetLogs(
		ctx context.Context,
		event brignext.Event,
		selector brignext.LogsSelector,
		opts meta.ListOptions,
	) (brignext.LogEntryList, error)
	StreamLogs(
		ctx context.Context,
		event brignext.Event,
		selector brignext.LogsSelector,
	) (<-chan brignext.LogEntry, error)
}
