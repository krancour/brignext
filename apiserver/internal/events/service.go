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

// Service is the specialized interface for managing Events. It's decoupled from
// underlying technology choices (e.g. data store, message bus, etc.) to keep
// business logic reusable and consistent while the underlying tech stack
// remains free to change.
type Service interface {
	// Create creates a new Event.
	Create(context.Context, brignext.Event) (
		brignext.EventReferenceList,
		error,
	)
	// List returns an EventReferenceList, with its EventReferences ordered by
	// age, newest first. Criteria for which Events should be retrieved can be
	// specified using the EventListOptions parameter.
	List(
		context.Context,
		brignext.EventListOptions,
	) (brignext.EventReferenceList, error)
	// Get retrieves a single Event specified by its identifier.
	Get(context.Context, string) (brignext.Event, error)
	// Cancel cancels a single Event specified by its identifier.
	Cancel(context.Context, string) error
	// CancelMany cancels multiple Events specified by the EventListOptions
	// parameter.
	CancelMany(
		context.Context,
		brignext.EventListOptions,
	) (brignext.EventReferenceList, error)
	// Delete deletes a single Event specified by its identifier.
	Delete(context.Context, string) error
	// DeleteMany deletes multiple Events specified by the EventListOptions
	// parameter.
	DeleteMany(
		context.Context,
		brignext.EventListOptions,
	) (brignext.EventReferenceList, error)

	// UpdateWorkerStatus updates the status of an Event's Worker.
	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		status brignext.WorkerStatus,
	) error

	// UpdateJobStatus, given an Event identifier and Job name, updates the status
	// of that Job.
	UpdateJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status brignext.JobStatus,
	) error

	// GetLogs retrieves logs for an Event's Worker, or using the LogOptions
	// parameter, a Job spawned by that Worker (or specific container thereof).
	GetLogs(
		ctx context.Context,
		eventID string,
		opts brignext.LogOptions,
	) (brignext.LogEntryList, error)
	// StreamLogs returns a channel over which logs for an Event's Worker, or
	// using the LogOptions parameter, a Job spawned by that Worker (or specific
	// container thereof), are streamed.
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

// NewService returns a specialized interface for managing Events.
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

// TODO: There's a lot of stuff that happens in this function that maybe we
// should defer until later-- like when the worker pod actually gets created.
func (s *service) Create(
	ctx context.Context,
	event brignext.Event,
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.EventReferenceList{}

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

	workerSpec := project.Spec.WorkerTemplate

	if workerSpec.WorkspaceSize == "" {
		workerSpec.WorkspaceSize = "10Gi"
	}

	if event.Git != nil {
		if workerSpec.Git == nil {
			workerSpec.Git = &brignext.WorkerGitConfig{}
		}
		// VCS details from the event override project-level details
		// TODO: Might need some nil checks below
		if event.Git.CloneURL != "" {
			workerSpec.Git.CloneURL = event.Git.CloneURL
		}
		if event.Git.Commit != "" {
			workerSpec.Git.Commit = event.Git.Commit
		}
		if event.Git.Ref != "" {
			workerSpec.Git.Ref = event.Git.Ref
		}
	}
	if workerSpec.Git != nil {
		if workerSpec.Git.CloneURL != "" &&
			workerSpec.Git.Commit == "" &&
			workerSpec.Git.Ref == "" {
			workerSpec.Git.Ref = "master"
		}
	}

	if workerSpec.LogLevel == "" {
		workerSpec.LogLevel = brignext.LogLevelInfo
	}

	if workerSpec.ConfigFilesDirectory == "" {
		workerSpec.ConfigFilesDirectory = "."
	}

	event.Worker = brignext.Worker{
		Spec: workerSpec,
		Status: brignext.WorkerStatus{
			Phase: brignext.WorkerPhasePending,
		},
	}

	// Let the scheduler add scheduler-specific details before we persist.
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
		return eventRefList, errors.Wrapf(
			err,
			"error storing new event %q",
			event.ID,
		)
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
	eventList := brignext.EventReferenceList{}

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

func (s *service) CancelMany(
	ctx context.Context,
	opts brignext.EventListOptions,
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.EventReferenceList{}

	// Refuse requests not qualified by project
	if opts.ProjectID == "" {
		return eventRefList, &brignext.ErrBadRequest{
			Reason: "Requests to cancel multiple events must be qualified by " +
				"project.",
		}
	}
	// Refuse requeets not qualified by worker phases
	if len(opts.WorkerPhases) == 0 {
		return eventRefList, &brignext.ErrBadRequest{
			Reason: "Requests to cancel multiple events must be qualified by " +
				"worker phase(s).",
		}
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

	eventRefList, err := s.store.CancelMany(ctx, opts)
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

func (s *service) DeleteMany(
	ctx context.Context,
	opts brignext.EventListOptions,
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.EventReferenceList{}

	// Refuse requests not qualified by project
	if opts.ProjectID == "" {
		return eventRefList, &brignext.ErrBadRequest{
			Reason: "Requests to delete multiple events must be qualified by " +
				"project.",
		}
	}
	// Refuse requeets not qualified by worker phases
	if len(opts.WorkerPhases) == 0 {
		return eventRefList, &brignext.ErrBadRequest{
			Reason: "Requests to delete multiple events must be qualified by " +
				"worker phase(s).",
		}
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

	eventRefList, err := s.store.DeleteMany(ctx, opts)
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
