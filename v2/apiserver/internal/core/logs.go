package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// LogsSelector represents useful criteria for selecting logs for streaming from
// a specific container of a Worker or Job.
type LogsSelector struct {
	// Job specifies, by name, a Job spawned by the Worker. If this field is
	// left blank, it is presumed logs are desired for the Worker itself.
	Job string
	// Container specifies, by name, a container belonging to the Worker or Job
	// whose logs are being retrieved. If left blank, a container with the same
	// name as the Worker or Job is assumed.
	Container string
}

// LogStreamOptions represents useful options for streaming logs from a specific
// container of a Worker or Job.
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

// LogsService is the specialized interface for managing Logs. It's
// decoupled from underlying technology choices (e.g. data store, message bus,
// etc.) to keep business logic reusable and consistent while the underlying
// tech stack remains free to change.
type LogsService interface {
	// Stream returns a channel over which logs for an Event's Worker, or using
	// the LogsSelector parameter, a Job spawned by that Worker (or specific
	// container thereof), are streamed. If the specified Event, Job, or Container
	// thereof does not exist, implementations MUST return a *meta.ErrNotFound
	// error.
	Stream(
		ctx context.Context,
		eventID string,
		selector LogsSelector,
		opts LogStreamOptions,
	) (<-chan LogEntry, error)
}

type logsService struct {
	authorize     authx.AuthorizeFn
	projectsStore ProjectsStore
	eventsStore   EventsStore
	warmLogsStore LogsStore
	coolLogsStore LogsStore
}

func NewLogsService(
	projectsStore ProjectsStore,
	eventsStore EventsStore,
	warmLogsStore LogsStore,
	coolLogsStore LogsStore,
) LogsService {
	return &logsService{
		authorize:     authx.Authorize,
		projectsStore: projectsStore,
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

	project, err := l.projectsStore.Get(ctx, event.ProjectID)
	if err != nil {
		return nil,
			errors.Wrapf(
				err,
				"error retrieving project %q from store",
				event.ProjectID,
			)
	}

	// Try warm logs first and fall back on cooler logs if necessary
	logCh, err := l.warmLogsStore.StreamLogs(ctx, project, event, selector, opts)
	if err != nil {
		logCh, err = l.coolLogsStore.StreamLogs(ctx, project, event, selector, opts)
	}
	return logCh, err
}

// LogsStore is an interface for components that implement Log persistence
// concerns.
type LogsStore interface {
	// Stream returns a channel over which logs for an Event's Worker, or using
	// the LogsSelector parameter, a Job spawned by that Worker (or specific
	// container thereof), are streamed. If the specified Event, Job, or Container
	// thereof does not exist, implementations MUST return a *meta.ErrNotFound
	// error.
	StreamLogs(
		ctx context.Context,
		project Project,
		event Event,
		selector LogsSelector,
		opts LogStreamOptions,
	) (<-chan LogEntry, error)
}
