package events

import (
	"context"
	"log"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/projects"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type Service interface {
	Create(context.Context, brignext.Event) (
		brignext.EventReferenceList,
		error,
	)
	List(
		context.Context,
		brignext.EventListOptions,
	) (brignext.EventReferenceList, error)
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
}

type service struct {
	projectsStore projects.Store
	store         Store
	logsStore     LogsStore
	scheduler     Scheduler
}

func NewService(
	projectsStore projects.Store,
	store Store,
	logsStore LogsStore,
	scheduler Scheduler,
) Service {
	return &service{
		projectsStore: projectsStore,
		store:         store,
		scheduler:     scheduler,
		logsStore:     logsStore,
	}
}

func (s *service) Create(
	ctx context.Context,
	event brignext.Event,
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.NewEventReferenceList()

	now := time.Now()
	event.Created = &now

	// If no project ID is specified, we use other criteria to locate projects
	// that are subscribed to this event. We iterate over all of those and create
	// an event for each of these by making a recursive call to this same
	// function.
	if event.ProjectID == "" {
		projectList, err := s.projectsStore.ListSubscribers(ctx, event)
		if err != nil {
			return eventRefList, errors.Wrap(
				err,
				"error retrieving subscribed projects from store",
			)
		}
		eventRefList.Items = make([]brignext.EventReference, len(projectList.Items))
		for i, project := range projectList.Items {
			event.ProjectID = project.ID
			eRefs, err := s.Create(ctx, event)
			if err != nil {
				return eventRefList, err
			}
			// eids will always contain precisely one element
			eventRefList.Items[i] = eRefs.Items[0]
		}
		return eventRefList, nil
	}

	// Make sure the project exists
	project, err := s.projectsStore.Get(ctx, event.ProjectID)
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

	// Let the scheduler add sheduler-specific details before we persist.
	if event, err = s.scheduler.PreCreate(ctx, project, event); err != nil {
		return eventRefList, errors.Wrapf(
			err,
			"error pre-creating event %q in scheduler",
			event.ID,
		)
	}

	// TODO: We'd like to use transaction semantics here, but transactions in
	// MongoDB are dicey, so we should refine this strategy to where a
	// partially completed create leaves us, overall, in a tolerable state.

	if err = s.store.Create(ctx, event); err != nil {
		return eventRefList, errors.Wrapf(err, "error storing new event %q", event.ID)
	}
	if err = s.scheduler.Create(ctx, project, event); err != nil {
		return eventRefList, errors.Wrapf(
			err,
			"error creating event %q in scheduler",
			event.ID,
		)
	}

	eventRefList.Items = []brignext.EventReference{
		brignext.NewEventReference(event),
	}
	return eventRefList, nil
}

func (s *service) List(
	ctx context.Context,
	opts brignext.EventListOptions,
) (brignext.EventReferenceList, error) {
	eventList := brignext.NewEventReferenceList()

	// If no worker phase filters were applied, retrieve all phases
	if len(opts.WorkerPhases) == 0 {
		opts.WorkerPhases = brignext.WorkerPhasesAll()
	}

	eventList, err := s.store.List(ctx, opts)
	if err != nil {
		return eventList, errors.Wrap(err, "error retrieving events from store")
	}
	return eventList, nil
}

func (s *service) Get(
	ctx context.Context,
	id string,
) (brignext.Event, error) {
	event, err := s.store.Get(ctx, id)
	if err != nil {
		return event, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			id,
		)
	}
	return event, nil
}

func (s *service) Cancel(ctx context.Context, id string) error {
	event, err := s.store.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	if err = s.store.Cancel(ctx, id); err != nil {
		return errors.Wrapf(err, "error canceling event %q in store", id)
	}

	go func() {
		if err = s.scheduler.Delete(
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

func (s *service) CancelCollection(
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
		_, err := s.store.Get(ctx, opts.ProjectID)
		if err != nil {
			return eventRefList, errors.Wrapf(
				err,
				"error retrieving project %q from store",
				opts.ProjectID,
			)
		}
	}

	eventRefList, err := s.store.CancelCollection(ctx, opts)
	if err != nil {
		return eventRefList, errors.Wrap(err, "error canceling events in store")
	}

	go func() {
		for _, eventRef := range eventRefList.Items {
			if err := s.scheduler.Delete(
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

func (s *service) Delete(ctx context.Context, id string) error {
	event, err := s.store.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	if err = s.store.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error deleting event %q from store", id)
	}

	go func() {
		if err = s.scheduler.Delete(
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

func (s *service) DeleteCollection(
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
		_, err := s.projectsStore.Get(ctx, opts.ProjectID)
		if err != nil {
			return eventRefList, errors.Wrapf(
				err,
				"error retrieving project %q from store",
				opts.ProjectID,
			)
		}
	}

	eventRefList, err := s.store.DeleteCollection(ctx, opts)
	if err != nil {
		return eventRefList, errors.Wrap(err, "error deleting events from store")
	}

	go func() {
		for _, eventRef := range eventRefList.Items {
			if err := s.scheduler.Delete(
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

func (s *service) UpdateWorkerStatus(
	ctx context.Context,
	eventID string,
	status brignext.WorkerStatus,
) error {
	if err := s.store.UpdateWorkerStatus(
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

func (s *service) UpdateJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status brignext.JobStatus,
) error {
	if err := s.store.UpdateJobStatus(
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

func (s *service) GetLogs(
	ctx context.Context,
	eventID string,
	opts brignext.LogOptions,
) (brignext.LogEntryList, error) {
	logEntryList := brignext.LogEntryList{}
	_, err := s.store.Get(ctx, eventID)
	if err != nil {
		return logEntryList, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			eventID,
		)
	}
	return s.logsStore.GetLogs(ctx, eventID, opts)
}

func (s *service) StreamLogs(
	ctx context.Context,
	eventID string,
	opts brignext.LogOptions,
) (<-chan brignext.LogEntry, error) {
	_, err := s.store.Get(ctx, eventID)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			eventID,
		)
	}
	return s.logsStore.StreamLogs(ctx, eventID, opts)
}
