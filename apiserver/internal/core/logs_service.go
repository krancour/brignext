package core

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/pkg/errors"
)

type LogsService interface {
	// Stream returns a channel over which logs for an Event's Worker, or
	// using the LogsSelector parameter, a Job spawned by that Worker (or specific
	// container thereof), are streamed.
	Stream(
		ctx context.Context,
		eventID string,
		selector LogsSelector,
		opts LogStreamOptions,
	) (<-chan LogEntry, error)
}

type logsService struct {
	authorize     authx.AuthorizeFn
	eventsStore   EventsStore
	warmLogsStore LogsStore
	coolLogsStore LogsStore
}

func NewLogsService(
	eventsStore EventsStore,
	warmLogsStore LogsStore,
	coolLogsStore LogsStore,
) LogsService {
	return &logsService{
		authorize:     authx.Authorize,
		eventsStore:   eventsStore,
		warmLogsStore: warmLogsStore,
		coolLogsStore: coolLogsStore,
	}
}

func (l *logsService) Stream(
	ctx context.Context,
	eventID string,
	selector LogsSelector,
	opts LogStreamOptions,
) (<-chan LogEntry, error) {
	event, err := l.eventsStore.Get(ctx, eventID)
	if err != nil {
		return nil,
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}

	if err := l.authorize(
		ctx,
		authx.RoleProjectUser(event.ProjectID),
	); err != nil {
		return nil, err
	}

	// Try warm logs first and fall back on cooler logs if necessary
	logCh, err := l.warmLogsStore.StreamLogs(ctx, event, selector, opts)
	if err != nil {
		logCh, err = l.coolLogsStore.StreamLogs(ctx, event, selector, opts)
	}
	return logCh, err
}
