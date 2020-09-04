package core

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/crypto"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// EventsService is the specialized interface for managing Events. It's decoupled from
// underlying technology choices (e.g. data store, message bus, etc.) to keep
// business logic reusable and consistent while the underlying tech stack
// remains free to change.
type EventsService interface {
	// Create creates a new Event.
	Create(context.Context, Event) (
		EventList,
		error,
	)
	// List returns an EventList, with its Items (Events) ordered by
	// age, newest first. Criteria for which Events should be retrieved can be
	// specified using the EventListOptions parameter.
	List(
		context.Context,
		EventsSelector,
		meta.ListOptions,
	) (EventList, error)
	// Get retrieves a single Event specified by its identifier.
	Get(context.Context, string) (Event, error)
	// GetByWorkerToken retrieves a single Event specified by its Worker's token.
	GetByWorkerToken(context.Context, string) (Event, error)
	// Cancel cancels a single Event specified by its identifier.
	Cancel(context.Context, string) error
	// CancelMany cancels multiple Events specified by the EventListOptions
	// parameter.
	CancelMany(
		context.Context,
		EventsSelector,
	) (CancelManyEventsResult, error)
	// Delete deletes a single Event specified by its identifier.
	Delete(context.Context, string) error
	// DeleteMany deletes multiple Events specified by the EventListOptions
	// parameter.
	DeleteMany(
		context.Context,
		EventsSelector,
	) (DeleteManyEventsResult, error)

	// StartWorker starts the indicated Event's Worker on BrigNext's workload
	// execution substrate.
	StartWorker(ctx context.Context, eventID string) error
	// GetWorkerStatus returns an Event's Worker's status.
	GetWorkerStatus(
		ctx context.Context,
		eventID string,
	) (WorkerStatus, error)
	// WatchWorkerStatus returns a channel over which an Event's Worker's status
	// is streamed. The channel receives a new WorkerStatus every time there is
	// any change in that status.
	WatchWorkerStatus(
		ctx context.Context,
		eventID string,
	) (<-chan WorkerStatus, error)
	// UpdateWorkerStatus updates the status of an Event's Worker.
	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		status WorkerStatus,
	) error

	// CreateJob, given an Event identifier and JobSpec, creates a new Job and
	// starts it on BrigNext's workload execution substrate.
	CreateJob(
		ctx context.Context,
		eventID string,
		jobName string,
		jobSpec JobSpec,
	) error
	// StartJob, given an Event identifier and Job name, starts that Job on
	// BrigNext's workload execution substrate.
	StartJob(
		ctx context.Context,
		eventID string,
		jobName string,
	) error
	// GetJobStatus, given an Event identifier and Job name, returns the Job's
	// status.
	GetJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
	) (JobStatus, error)
	// WatchJobStatus, given an Event identifier and Job name, returns a channel
	// over which the Job's status is streamed. The channel receives a new
	// JobStatus every time there is any change in that status.
	WatchJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan JobStatus, error)
	// UpdateJobStatus, given an Event identifier and Job name, updates the status
	// of that Job.
	UpdateJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status JobStatus,
	) error

	// StreamLogs returns a channel over which logs for an Event's Worker, or
	// using the LogsSelector parameter, a Job spawned by that Worker (or specific
	// container thereof), are streamed.
	StreamLogs(
		ctx context.Context,
		eventID string,
		selector LogsSelector,
		opts LogStreamOptions,
	) (<-chan LogEntry, error)
}

type eventsService struct {
	authorize     authx.AuthorizeFn
	projectsStore ProjectsStore
	store         EventsStore
	warmLogsStore LogsStore
	coolLogsStore LogsStore
	scheduler     EventsScheduler
}

// NewEventsService returns a specialized interface for managing Events.
func NewEventsService(
	projectsStore ProjectsStore,
	store EventsStore,
	warmLogsStore LogsStore,
	coolLogsStore LogsStore,
	scheduler EventsScheduler,
) EventsService {
	return &eventsService{
		authorize:     authx.Authorize,
		projectsStore: projectsStore,
		store:         store,
		scheduler:     scheduler,
		warmLogsStore: warmLogsStore,
		coolLogsStore: coolLogsStore,
	}
}

// TODO: There's a lot of stuff that happens in this function that maybe we
// should defer until later-- like when the worker pod actually gets created.
func (e *eventsService) Create(
	ctx context.Context,
	event Event,
) (EventList, error) {
	events := EventList{}

	if event.ProjectID == "" {
		if err := e.authorize(
			ctx,
			authx.RoleEventCreator(event.Source),
		); err != nil {
			return events, err
		}
	} else {
		if err := e.authorize(
			ctx,
			authx.RoleProjectUser(event.ProjectID),
			authx.RoleEventCreator(event.Source),
		); err != nil {
			return events, err
		}
	}

	now := time.Now()
	event.Created = &now

	// If no project ID is specified, we use other criteria to locate projects
	// that are subscribed to this event. We iterate over all of those and create
	// an event for each of these by making a recursive call to this same
	// function.
	if event.ProjectID == "" {
		projects, err := e.projectsStore.ListSubscribers(ctx, event)
		if err != nil {
			return events, errors.Wrap(
				err,
				"error retrieving subscribed projects from store",
			)
		}
		events.Items = make([]Event, len(projects.Items))
		for i, project := range projects.Items {
			event.ProjectID = project.ID
			projectEvents, err := e.Create(ctx, event)
			if err != nil {
				return events, err
			}
			// projectEvents.Items will always contain precisely one element
			events.Items[i] = projectEvents.Items[0]
		}
		return events, nil
	}

	// Make sure the project exists
	project, err := e.projectsStore.Get(ctx, event.ProjectID)
	if err != nil {
		return events, errors.Wrapf(
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
			workerSpec.Git = &WorkerGitConfig{}
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
		workerSpec.LogLevel = LogLevelInfo
	}

	if workerSpec.ConfigFilesDirectory == "" {
		workerSpec.ConfigFilesDirectory = "."
	}

	token := crypto.NewToken(256)

	event.Worker = Worker{
		Spec: workerSpec,
		Status: WorkerStatus{
			Phase: WorkerPhasePending,
		},
		Token:       token,
		HashedToken: crypto.ShortSHA("", token),
	}

	// Let the scheduler add scheduler-specific details before we persist.
	if event, err = e.scheduler.PreCreate(ctx, project, event); err != nil {
		return events, errors.Wrapf(
			err,
			"error pre-creating event %q in scheduler",
			event.ID,
		)
	}

	if err = e.store.Create(ctx, event); err != nil {
		return events, errors.Wrapf(
			err,
			"error storing new event %q",
			event.ID,
		)
	}
	if err = e.scheduler.Create(ctx, project, event); err != nil {
		return events, errors.Wrapf(
			err,
			"error creating event %q in scheduler",
			event.ID,
		)
	}

	events.Items = []Event{event}
	return events, nil
}

func (e *eventsService) List(
	ctx context.Context,
	selector EventsSelector,
	opts meta.ListOptions,
) (EventList, error) {
	if err := e.authorize(ctx, authx.RoleReader()); err != nil {
		return EventList{}, err
	}

	// If no worker phase filters were applied, retrieve all phases
	if len(selector.WorkerPhases) == 0 {
		selector.WorkerPhases = WorkerPhasesAll()
	}
	if opts.Limit == 0 {
		opts.Limit = 20
	}

	events, err := e.store.List(ctx, selector, opts)
	if err != nil {
		return events, errors.Wrap(err, "error retrieving events from store")
	}
	return events, nil
}

func (e *eventsService) Get(
	ctx context.Context,
	id string,
) (Event, error) {
	if err := e.authorize(ctx, authx.RoleReader()); err != nil {
		return Event{}, err
	}

	event, err := e.store.Get(ctx, id)
	if err != nil {
		return event, errors.Wrapf(err, "error retrieving event %q from store", id)
	}
	return event, nil
}

func (e *eventsService) GetByWorkerToken(
	ctx context.Context,
	workerToken string,
) (Event, error) {
	// No authz is required here because this is only ever called by the system
	// itself.

	event, err := e.store.GetByHashedWorkerToken(
		ctx,
		crypto.ShortSHA("", workerToken),
	)
	if err != nil {
		return event, errors.Wrap(err, "error retrieving event from store")
	}
	return event, nil
}

func (e *eventsService) Cancel(ctx context.Context, id string) error {
	event, err := e.store.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	if err = e.authorize(
		ctx,
		authx.RoleProjectUser(event.ProjectID),
	); err != nil {
		return err
	}

	if err = e.store.Cancel(ctx, id); err != nil {
		return errors.Wrapf(err, "error canceling event %q in store", id)
	}

	go func() {
		if err = e.scheduler.Delete(
			context.Background(), // Deliberately not using request context
			event,
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

func (e *eventsService) CancelMany(
	ctx context.Context,
	selector EventsSelector,
) (CancelManyEventsResult, error) {
	result := CancelManyEventsResult{}

	// Refuse requests not qualified by project
	if selector.ProjectID == "" {
		return result, &meta.ErrBadRequest{
			Reason: "Requests to cancel multiple events must be qualified by " +
				"project.",
		}
	}

	if err := e.authorize(
		ctx,
		authx.RoleProjectUser(selector.ProjectID),
	); err != nil {
		return CancelManyEventsResult{}, err
	}

	// Refuse requests not qualified by worker phases
	if len(selector.WorkerPhases) == 0 {
		return result, &meta.ErrBadRequest{
			Reason: "Requests to cancel multiple events must be qualified by " +
				"worker phase(s).",
		}
	}

	if selector.ProjectID != "" {
		// Make sure the project exists
		_, err := e.projectsStore.Get(ctx, selector.ProjectID)
		if err != nil {
			return result, errors.Wrapf(
				err,
				"error retrieving project %q from store",
				selector.ProjectID,
			)
		}
	}

	events, err := e.store.CancelMany(ctx, selector)
	if err != nil {
		return result, errors.Wrap(err, "error canceling events in store")
	}

	result.Count = int64(len(events.Items))

	// TODO: Can we find a quicker, more efficient way to do this?
	go func() {
		for _, event := range events.Items {
			if err := e.scheduler.Delete(
				context.Background(), // Deliberately not using request context
				event,
			); err != nil {
				log.Println(
					errors.Wrapf(
						err,
						"error deleting event %q from scheduler",
						event.ID,
					),
				)
			}
		}
	}()

	return result, nil
}

func (e *eventsService) Delete(ctx context.Context, id string) error {
	event, err := e.store.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	if err = e.authorize(
		ctx,
		authx.RoleProjectUser(event.ProjectID),
	); err != nil {
		return err
	}

	if err = e.store.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error deleting event %q from store", id)
	}

	go func() {
		if err = e.scheduler.Delete(
			context.Background(), // Deliberately not using request context
			event,
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

func (e *eventsService) DeleteMany(
	ctx context.Context,
	selector EventsSelector,
) (DeleteManyEventsResult, error) {
	result := DeleteManyEventsResult{}

	// Refuse requests not qualified by project
	if selector.ProjectID == "" {
		return result, &meta.ErrBadRequest{
			Reason: "Requests to delete multiple events must be qualified by " +
				"project.",
		}
	}

	if err := e.authorize(
		ctx,
		authx.RoleProjectUser(selector.ProjectID),
	); err != nil {
		return DeleteManyEventsResult{}, err
	}

	// Refuse requests not qualified by worker phases
	if len(selector.WorkerPhases) == 0 {
		return result, &meta.ErrBadRequest{
			Reason: "Requests to delete multiple events must be qualified by " +
				"worker phase(s).",
		}
	}

	if selector.ProjectID != "" {
		// Make sure the project exists
		_, err := e.projectsStore.Get(ctx, selector.ProjectID)
		if err != nil {
			return result, errors.Wrapf(
				err,
				"error retrieving project %q from store",
				selector.ProjectID,
			)
		}
	}

	events, err := e.store.DeleteMany(ctx, selector)
	if err != nil {
		return result, errors.Wrap(err, "error deleting events from store")
	}

	result.Count = int64(len(events.Items))

	// TODO: Can we find a quicker, more efficient way to do this?
	go func() {
		for _, event := range events.Items {
			if err := e.scheduler.Delete(
				context.Background(), // Deliberately not using request context
				event,
			); err != nil {
				log.Println(
					errors.Wrapf(
						err,
						"error deleting event %q from scheduler",
						event.ID,
					),
				)
			}
		}
	}()

	return result, nil
}

func (e *eventsService) StartWorker(ctx context.Context, eventID string) error {
	if err := e.authorize(ctx, authx.RoleScheduler()); err != nil {
		return err
	}

	event, err := e.store.Get(ctx, eventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}

	// spec := event.Worker.Spec
	// TODO: This is probably a better place to apply worker default just before
	// it is started INSTEAD OF setting defaults at event creation time or waiting
	// all the way up until pod creation time.
	// if err = e.store.UpdateWorkerSpec(ctx, eventID, spec); err != nil {
	// 	return errors.Wrapf(
	// 		err,
	// 		"error updating worker's spec for event %q",
	// 		event.ID,
	// 	)
	// }

	if event.Worker.Status.Phase != WorkerPhasePending {
		return &meta.ErrConflict{
			Type: "Event",
			ID:   event.ID,
			Reason: fmt.Sprintf(
				"Event %q worker has already been started.",
				event.ID,
			),
		}
	}

	if err = e.scheduler.StartWorker(ctx, event); err != nil {
		return errors.Wrapf(err, "error starting worker for event %q", event.ID)
	}
	return nil
}

func (e *eventsService) GetWorkerStatus(
	ctx context.Context,
	eventID string,
) (WorkerStatus, error) {
	if err := e.authorize(ctx, authx.RoleReader()); err != nil {
		return WorkerStatus{}, err
	}

	event, err := e.store.Get(ctx, eventID)
	if err != nil {
		return WorkerStatus{},
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	return event.Worker.Status, nil
}

// TODO: Should we put some kind of timeout on this function?
func (e *eventsService) WatchWorkerStatus(
	ctx context.Context,
	eventID string,
) (<-chan WorkerStatus, error) {
	if err := e.authorize(ctx, authx.RoleReader()); err != nil {
		return nil, err
	}

	// Read the event up front to confirm it exists.
	if _, err := e.store.Get(ctx, eventID); err != nil {
		return nil,
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	statusCh := make(chan WorkerStatus)
	go func() {
		defer close(statusCh)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
			event, err := e.store.Get(ctx, eventID)
			if err != nil {
				log.Printf("error retrieving event %q from store: %s", eventID, err)
				return
			}
			select {
			case statusCh <- event.Worker.Status:
			case <-ctx.Done():
				return
			}
		}
	}()
	return statusCh, nil
}

func (e *eventsService) UpdateWorkerStatus(
	ctx context.Context,
	eventID string,
	status WorkerStatus,
) error {
	if err := e.authorize(ctx, authx.RoleObserver()); err != nil {
		return err
	}

	if err := e.store.UpdateWorkerStatus(
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

func (e *eventsService) CreateJob(
	ctx context.Context,
	eventID string,
	jobName string,
	jobSpec JobSpec,
) error {
	if err := e.authorize(ctx, authx.RoleWorker(eventID)); err != nil {
		return err
	}

	event, err := e.store.Get(ctx, eventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	if _, ok := event.Worker.Jobs[jobName]; ok {
		return &meta.ErrConflict{
			Type: "Job",
			ID:   jobName,
			Reason: fmt.Sprintf(
				"Event %q already has a job named %q.",
				eventID,
				jobName,
			),
		}
	}

	// Perform some validations...

	// Determine if ANY of the job's containers:
	//   1. Use shared workspace
	//   2. Run in privileged mode
	//   3. Mount the host's Docker socket
	var useWorkspace = jobSpec.PrimaryContainer.UseWorkspace
	var usePrivileged = jobSpec.PrimaryContainer.Privileged
	var useDockerSocket = jobSpec.PrimaryContainer.UseHostDockerSocket
	for _, sidecarContainer := range jobSpec.SidecarContainers {
		if sidecarContainer.UseWorkspace {
			useWorkspace = true
		}
		if sidecarContainer.Privileged {
			usePrivileged = true
		}
		if sidecarContainer.UseHostDockerSocket {
			useDockerSocket = true
		}
	}

	// Fail quickly if any job is trying to run privileged or use the host's
	// Docker socket, but isn't allowed to per worker configuration.
	if usePrivileged &&
		(event.Worker.Spec.JobPolicies == nil ||
			!event.Worker.Spec.JobPolicies.AllowPrivileged) {
		return &meta.ErrAuthorization{
			Reason: "Worker configuration forbids jobs from utilizing privileged " +
				"containers.",
		}
	}
	if useDockerSocket &&
		(event.Worker.Spec.JobPolicies == nil ||
			!event.Worker.Spec.JobPolicies.AllowDockerSocketMount) {
		return &meta.ErrAuthorization{
			Reason: "Worker configuration forbids jobs from mounting the Docker " +
				"socket.",
		}
	}

	// Fail quickly if the job needs to use shared workspace, but the worker
	// doesn't have any shared workspace.
	if useWorkspace && !event.Worker.Spec.UseWorkspace {
		return &meta.ErrConflict{
			Reason: "The job requested access to the shared workspace, but Worker " +
				"configuration has not enabled this feature.",
		}
	}

	if err = e.store.CreateJob(ctx, eventID, jobName, jobSpec); err != nil {
		return errors.Wrapf(
			err, "error saving event %q job %q in store",
			eventID,
			eventID,
		)
	}

	if err = e.scheduler.CreateJob(ctx, event, jobName); err != nil {
		return errors.Wrapf(
			err,
			"error scheduling event %q job %q",
			event.ID,
			jobName,
		)
	}

	return nil
}

func (e *eventsService) StartJob(
	ctx context.Context,
	eventID string,
	jobName string,
) error {
	if err := e.authorize(ctx, authx.RoleScheduler()); err != nil {
		return err
	}

	event, err := e.store.Get(ctx, eventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	job, ok := event.Worker.Jobs[jobName]
	if !ok {
		return &meta.ErrNotFound{
			Type: "Job",
			ID:   jobName,
		}
	}

	if job.Status.Phase != JobPhasePending {
		return &meta.ErrConflict{
			Type: "Job",
			ID:   jobName,
			Reason: fmt.Sprintf(
				"Event %q job %q has already been started.",
				eventID,
				jobName,
			),
		}
	}

	if err = e.scheduler.StartJob(ctx, event, jobName); err != nil {
		return errors.Wrapf(
			err,
			"error starting event %q job %q",
			event.ID,
			jobName,
		)
	}

	return nil
}

func (e *eventsService) GetJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (JobStatus, error) {
	if err := e.authorize(ctx, authx.RoleReader()); err != nil {
		return JobStatus{}, err
	}

	event, err := e.store.Get(ctx, eventID)
	if err != nil {
		return JobStatus{},
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	job, ok := event.Worker.Jobs[jobName]
	if !ok {
		return JobStatus{}, &meta.ErrNotFound{
			Type: "Job",
			ID:   jobName,
		}
	}
	return job.Status, nil
}

// TODO: Should we put some kind of timeout on this function?
func (e *eventsService) WatchJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan JobStatus, error) {
	if err := e.authorize(ctx, authx.RoleReader()); err != nil {
		return nil, err
	}

	// Read the event and job up front to confirm they both exists.
	event, err := e.store.Get(ctx, eventID)
	if err != nil {
		return nil,
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	if _, ok := event.Worker.Jobs[jobName]; !ok {
		return nil, &meta.ErrNotFound{
			Type: "Job",
			ID:   jobName,
		}
	}
	statusCh := make(chan JobStatus)
	go func() {
		defer close(statusCh)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
			event, err := e.store.Get(ctx, eventID)
			if err != nil {
				log.Printf("error retrieving event %q from store: %s", eventID, err)
				return
			}
			select {
			case statusCh <- event.Worker.Jobs[jobName].Status:
			case <-ctx.Done():
				return
			}
		}
	}()
	return statusCh, nil
}

func (e *eventsService) UpdateJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status JobStatus,
) error {
	if err := e.authorize(ctx, authx.RoleObserver()); err != nil {
		return err
	}

	if err := e.store.UpdateJobStatus(
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

func (e *eventsService) StreamLogs(
	ctx context.Context,
	eventID string,
	selector LogsSelector,
	opts LogStreamOptions,
) (<-chan LogEntry, error) {
	if err := e.authorize(ctx, authx.RoleReader()); err != nil {
		return nil, err
	}

	event, err := e.store.Get(ctx, eventID)
	if err != nil {
		return nil,
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	// Try warm logs first and fall back on cooler logs if necessary
	logCh, err := e.warmLogsStore.StreamLogs(ctx, event, selector, opts)
	if err != nil {
		logCh, err = e.coolLogsStore.StreamLogs(ctx, event, selector, opts)
	}
	return logCh, err
}
