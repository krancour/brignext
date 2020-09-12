package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// LogsSelector represents useful criteria when requesting a log stream from an
// Event.
type LogsSelector struct {
	// Job specifies, by name, a Job spawned by the Worker. If this field is
	// left blank, it is presumed logs are desired for the Worker itself.
	Job string
	// Container specifies, by name, a container belonging to the Worker or Job
	// whose logs are being retrieved. If left blank, a container with the same
	// name as the Worker or Job is assumed.
	Container string
}

// LogStreamOptions represents useful options when requesting a log stream from
// an Event.
type LogStreamOptions struct {
	// Follow indicates whether the stream should conclude after the last
	// available line of logs has been sent to the client (false) or remain open
	// until closed by the client (true), continuing to send new lines as they
	// become available.
	Follow bool `json:"follow"`
}

// LogEntry represents one line of output from an OCI container.
type LogEntry struct {
	// Time is the time the line was written.
	Time *time.Time `json:"time,omitempty" bson:"time,omitempty"`
	// Message is a single line of log output from an OCI container.
	Message string `json:"message,omitempty" bson:"log,omitempty"`
}

// MarshalJSON amends LogEntry instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (l LogEntry) MarshalJSON() ([]byte, error) {
	type Alias LogEntry
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "LogEntry",
			},
			Alias: (Alias)(l),
		},
	)
}

// TODO: We probably don't need this interface. The idea is to have a single
// implementation of the service's logic, with only underlying components being
// pluggable.
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

// TODO: There probably isn't any good reason to actually have this
// constructor-like function here. Let's consider removing it.
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

	if err = l.authorize(
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

type LogsStore interface {
	StreamLogs(
		ctx context.Context,
		event Event,
		selector LogsSelector,
		opts LogStreamOptions,
	) (<-chan LogEntry, error)
}
