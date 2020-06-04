package service

import (
	"context"
	"log"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/scheduler"
	"github.com/krancour/brignext/v2/internal/apiserver/storage"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type EventsService interface {
	Create(context.Context, brignext.Event) (
		brignext.EventReferenceList,
		error,
	)
	List(context.Context, brignext.EventListOptions) (brignext.EventList, error)
	Get(context.Context, string) (brignext.Event, error)
	Cancel(context.Context, string) error
	CancelCollection(
		context.Context,
		brignext.EventListOptions,
	) (brignext.EventReferenceList, error)
	Delete(context.Context, string) error
	DeleteCollection(
		context.Context,
		brignext.EventListOptions,
	) (brignext.EventReferenceList, error)

	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		status brignext.WorkerStatus,
	) error

	UpdateJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status brignext.JobStatus,
	) error

	GetWorkerLogs(
		ctx context.Context,
		eventID string,
	) (brignext.LogEntryList, error)
	StreamWorkerLogs(
		ctx context.Context,
		eventID string,
	) (<-chan brignext.LogEntry, error)
	GetWorkerInitLogs(
		ctx context.Context,
		eventID string,
	) (brignext.LogEntryList, error)
	StreamWorkerInitLogs(
		ctx context.Context,
		eventID string,
	) (<-chan brignext.LogEntry, error)

	GetJobLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (brignext.LogEntryList, error)
	StreamJobLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan brignext.LogEntry, error)
	GetJobInitLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (brignext.LogEntryList, error)
	StreamJobInitLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan brignext.LogEntry, error)
}

type eventsService struct {
	store     storage.Store
	scheduler scheduler.EventsScheduler
	logStore  storage.LogsStore
}

func NewEventsService(
	store storage.Store,
	scheduler scheduler.EventsScheduler,
	logStore storage.LogsStore,
) EventsService {
	return &eventsService{
		store:     store,
		scheduler: scheduler,
		logStore:  logStore,
	}
}

func (e *eventsService) Create(
	ctx context.Context,
	event brignext.Event,
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.NewEventReferenceList()

	// If no project ID is specified, we use other criteria to locate projects
	// that are subscribed to this event. We iterate over all of those and create
	// an event for each of these by making a recursive call to this same
	// function.
	if event.ProjectID == "" {
		projectList, err := e.store.Projects().ListSubscribed(ctx, event)
		if err != nil {
			return eventRefList, errors.Wrap(
				err,
				"error retrieving subscribed projects from store",
			)
		}
		eventRefList.Items = make([]brignext.EventReference, len(projectList.Items))
		for i, project := range projectList.Items {
			event.ProjectID = project.ID
			eRefs, err := e.Create(ctx, event)
			if err != nil {
				return eventRefList, err
			}
			// eids will always contain precisely one element
			eventRefList.Items[i] = eRefs.Items[0]
		}
		return eventRefList, nil
	}

	// Make sure the project exists
	project, err := e.store.Projects().Get(ctx, event.ProjectID)
	if err != nil {
		return eventRefList, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			event.ProjectID,
		)
	}

	event.ID = uuid.NewV4().String()

	event.Worker = &project.Spec.Worker

	if event.Worker.WorkspaceSize == "" {
		event.Worker.WorkspaceSize = "10Gi"
	}

	// VCS details from the event override project-level details
	if event.Git.CloneURL != "" {
		event.Worker.Git.CloneURL = event.Git.CloneURL
	}
	if event.Git.Commit != "" {
		event.Worker.Git.Commit = event.Git.Commit
	}
	if event.Git.Ref != "" {
		event.Worker.Git.Ref = event.Git.Ref
	}

	if event.Worker.Git.CloneURL != "" &&
		event.Worker.Git.Commit == "" &&
		event.Worker.Git.Ref == "" {
		event.Worker.Git.Ref = "master"
	}

	if event.Worker.LogLevel == "" {
		event.Worker.LogLevel = brignext.LogLevelInfo
	}

	if event.Worker.ConfigFilesDirectory == "" {
		event.Worker.ConfigFilesDirectory = "."
	}

	event.Status = &brignext.EventStatus{
		WorkerStatus: brignext.WorkerStatus{
			Phase: brignext.WorkerPhasePending,
		},
		JobStatuses: map[string]brignext.JobStatus{},
	}

	// We send this to the scheduler first because we expect the scheduler will
	// will add some scheduler-specific details that we will want to persist.
	event, err = e.scheduler.Create(
		ctx,
		project,
		event,
	)
	if err != nil {
		return eventRefList, errors.Wrapf(
			err,
			"error creating event %q in scheduler",
			event.ID,
		)
	}
	if err := e.store.Events().Create(ctx, event); err != nil {
		// We need to roll this back manually because the scheduler doesn't
		// automatically roll anything back upon failure.
		// nolint: errcheck
		e.scheduler.Delete(ctx, brignext.EventReference{})
		return eventRefList,
			errors.Wrapf(err, "error storing new event %q", event.ID)
	}

	eventRefList.Items = []brignext.EventReference{
		brignext.NewEventReference(event),
	}
	return eventRefList, nil
}

func (e *eventsService) List(
	ctx context.Context,
	opts brignext.EventListOptions,
) (brignext.EventList, error) {
	eventList := brignext.NewEventList()

	// If no worker phase filters were applied, retrieve all phases
	if len(opts.WorkerPhases) == 0 {
		opts.WorkerPhases = brignext.WorkerPhasesAll()
	}

	eventList, err := e.store.Events().List(ctx, opts)
	if err != nil {
		return eventList, errors.Wrap(err, "error retrieving events from store")
	}
	return eventList, nil
}

func (e *eventsService) Get(
	ctx context.Context,
	id string,
) (brignext.Event, error) {
	event, err := e.store.Events().Get(ctx, id)
	if err != nil {
		return event, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			id,
		)
	}
	return event, nil
}

func (e *eventsService) Cancel(ctx context.Context, id string) error {
	event, err := e.store.Events().Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	if err = e.store.Events().Cancel(ctx, id); err != nil {
		return errors.Wrapf(err, "error canceling event %q in store", id)
	}

	go func() {
		if err = e.scheduler.Delete(
			context.Background(), // Deliberately not using request context
			brignext.NewEventReference(event),
		); err != nil {
			log.Println(
				errors.Wrapf(
					err,
					"error deleting event %q from scheduler",
					id,
				),
			)
		}
	}()

	return nil
}

func (e *eventsService) CancelCollection(
	ctx context.Context,
	opts brignext.EventListOptions,
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.NewEventReferenceList()

	// Refuse requests not qualified by project
	if opts.ProjectID == "" {
		return eventRefList, brignext.NewErrBadRequest(
			"Requests to cancel multiple events must be qualified by project.",
		)
	}
	// Refuse requeets not qualified by worker phases
	if len(opts.WorkerPhases) == 0 {
		return eventRefList, brignext.NewErrBadRequest(
			"Requests to cancel multiple events must be qualified by worker " +
				"phase(s).",
		)
	}

	if opts.ProjectID != "" {
		// Make sure the project exists
		_, err := e.store.Projects().Get(ctx, opts.ProjectID)
		if err != nil {
			return eventRefList, errors.Wrapf(
				err,
				"error retrieving project %q from store",
				opts.ProjectID,
			)
		}
	}

	eventRefList, err := e.store.Events().CancelCollection(ctx, opts)
	if err != nil {
		return eventRefList, errors.Wrap(err, "error canceling events in store")
	}

	go func() {
		for _, eventRef := range eventRefList.Items {
			if err := e.scheduler.Delete(
				context.Background(), // Deliberately not using request context
				eventRef,
			); err != nil {
				log.Println(
					errors.Wrapf(
						err,
						"error deleting event %q from scheduler",
						eventRef.ID,
					),
				)
			}
		}
	}()

	return eventRefList, nil
}

func (e *eventsService) Delete(ctx context.Context, id string) error {
	event, err := e.store.Events().Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	if err = e.store.Events().Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error deleting event %q from store", id)
	}

	go func() {
		if err = e.scheduler.Delete(
			context.Background(), // Deliberately not using request context
			brignext.NewEventReference(event),
		); err != nil {
			log.Println(
				errors.Wrapf(
					err,
					"error deleting event %q from scheduler",
					id,
				),
			)
		}
	}()

	return nil
}

func (e *eventsService) DeleteCollection(
	ctx context.Context,
	opts brignext.EventListOptions,
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.NewEventReferenceList()

	// Refuse requests not qualified by project
	if opts.ProjectID == "" {
		return eventRefList, brignext.NewErrBadRequest(
			"Requests to delete multiple events must be qualified by project.",
		)
	}
	// Refuse requeets not qualified by worker phases
	if len(opts.WorkerPhases) == 0 {
		return eventRefList, brignext.NewErrBadRequest(
			"Requests to delete multiple events must be qualified by worker " +
				"phase(s).",
		)
	}

	if opts.ProjectID != "" {
		// Make sure the project exists
		_, err := e.store.Projects().Get(ctx, opts.ProjectID)
		if err != nil {
			return eventRefList, errors.Wrapf(
				err,
				"error retrieving project %q from store",
				opts.ProjectID,
			)
		}
	}

	eventRefList, err := e.store.Events().DeleteCollection(ctx, opts)
	if err != nil {
		return eventRefList, errors.Wrap(err, "error deleting events from store")
	}

	go func() {
		for _, eventRef := range eventRefList.Items {
			if err := e.scheduler.Delete(
				context.Background(), // Deliberately not using request context
				eventRef,
			); err != nil {
				log.Println(
					errors.Wrapf(
						err,
						"error deleting event %q from scheduler",
						eventRef.ID,
					),
				)
			}
		}
	}()

	return eventRefList, nil
}

func (e *eventsService) UpdateWorkerStatus(
	ctx context.Context,
	eventID string,
	status brignext.WorkerStatus,
) error {
	if err := e.store.Events().UpdateWorkerStatus(
		ctx,
		eventID,
		status,
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker in store",
			eventID,
		)
	}
	return nil
}

func (e *eventsService) UpdateJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status brignext.JobStatus,
) error {
	if err := e.store.Events().UpdateJobStatus(
		ctx,
		eventID,
		jobName,
		status,
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker job %q in store",
			eventID,
			jobName,
		)
	}
	return nil
}

func (e *eventsService) GetWorkerLogs(
	ctx context.Context,
	eventID string,
) (brignext.LogEntryList, error) {
	logEntryList := brignext.LogEntryList{}
	_, err := e.store.Events().Get(ctx, eventID)
	if err != nil {
		return logEntryList, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			eventID,
		)
	}
	return e.logStore.GetWorkerLogs(ctx, eventID)
}

func (e *eventsService) StreamWorkerLogs(
	ctx context.Context,
	eventID string,
) (<-chan brignext.LogEntry, error) {
	_, err := e.store.Events().Get(ctx, eventID)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			eventID,
		)
	}
	return e.logStore.StreamWorkerLogs(ctx, eventID)
}

func (e *eventsService) GetWorkerInitLogs(
	ctx context.Context,
	eventID string,
) (brignext.LogEntryList, error) {
	logEntryList := brignext.LogEntryList{}
	_, err := e.store.Events().Get(ctx, eventID)
	if err != nil {
		return logEntryList, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			eventID,
		)
	}
	return e.logStore.GetWorkerInitLogs(ctx, eventID)
}

func (e *eventsService) StreamWorkerInitLogs(
	ctx context.Context,
	eventID string,
) (<-chan brignext.LogEntry, error) {
	_, err := e.store.Events().Get(ctx, eventID)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			eventID,
		)
	}
	return e.logStore.StreamWorkerInitLogs(ctx, eventID)
}

func (e *eventsService) GetJobLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (brignext.LogEntryList, error) {
	logEntryList := brignext.LogEntryList{}
	event, err := e.store.Events().Get(ctx, eventID)
	if err != nil {
		return logEntryList, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			eventID,
		)
	}
	if _, ok := event.Status.JobStatuses[jobName]; !ok {
		return logEntryList, brignext.NewErrNotFound("Job", jobName)
	}
	return e.logStore.GetJobLogs(ctx, eventID, jobName)
}

func (e *eventsService) StreamJobLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan brignext.LogEntry, error) {
	event, err := e.store.Events().Get(ctx, eventID)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			eventID,
		)
	}
	if _, ok := event.Status.JobStatuses[jobName]; !ok {
		return nil, brignext.NewErrNotFound("Job", jobName)
	}
	return e.logStore.StreamJobLogs(ctx, eventID, jobName)
}

func (e *eventsService) GetJobInitLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (brignext.LogEntryList, error) {
	logEntryList := brignext.LogEntryList{}
	event, err := e.store.Events().Get(ctx, eventID)
	if err != nil {
		return logEntryList, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			eventID,
		)
	}
	if _, ok := event.Status.JobStatuses[jobName]; !ok {
		return logEntryList, brignext.NewErrNotFound("Job", jobName)
	}
	return e.logStore.GetJobInitLogs(ctx, eventID, jobName)
}

func (e *eventsService) StreamJobInitLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan brignext.LogEntry, error) {
	event, err := e.store.Events().Get(ctx, eventID)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			eventID,
		)
	}
	if _, ok := event.Status.JobStatuses[jobName]; !ok {
		return nil, brignext.NewErrNotFound("Job", jobName)
	}
	return e.logStore.StreamJobInitLogs(ctx, eventID, jobName)
}
