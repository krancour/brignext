package events

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/crypto"
	"github.com/krancour/brignext/v2/apiserver/internal/projects"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
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
		brignext.EventList,
		error,
	)
	// List returns an EventList, with its Items (Events) ordered by
	// age, newest first. Criteria for which Events should be retrieved can be
	// specified using the EventListOptions parameter.
	List(
		context.Context,
		brignext.EventsSelector,
		meta.ListOptions,
	) (brignext.EventList, error)
	// Get retrieves a single Event specified by its identifier.
	Get(context.Context, string) (brignext.Event, error)
	// GetByWorkerToken retrieves a single Event specified by its Worker's token.
	GetByWorkerToken(context.Context, string) (brignext.Event, error)
	// Cancel cancels a single Event specified by its identifier.
	Cancel(context.Context, string) error
	// CancelMany cancels multiple Events specified by the EventListOptions
	// parameter.
	CancelMany(
		context.Context,
		brignext.EventsSelector,
	) (brignext.CancelManyEventsResult, error)
	// Delete deletes a single Event specified by its identifier.
	Delete(context.Context, string) error
	// DeleteMany deletes multiple Events specified by the EventListOptions
	// parameter.
	DeleteMany(
		context.Context,
		brignext.EventsSelector,
	) (brignext.DeleteManyEventsResult, error)

	// StartWorker starts the indicated Event's Worker on BrigNext's workload
	// execution substrate.
	StartWorker(ctx context.Context, eventID string) error
	// GetWorkerStatus returns an Event's Worker's status.
	GetWorkerStatus(
		ctx context.Context,
		eventID string,
	) (brignext.WorkerStatus, error)
	// WatchWorkerStatus returns a channel over which an Event's Worker's status
	// is streamed. The channel receives a new WorkerStatus every time there is
	// any change in that status.
	WatchWorkerStatus(
		ctx context.Context,
		eventID string,
	) (<-chan brignext.WorkerStatus, error)
	// UpdateWorkerStatus updates the status of an Event's Worker.
	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		status brignext.WorkerStatus,
	) error

	// CreateJob, given an Event identifier and JobSpec, creates a new Job and
	// starts it on BrigNext's worlkoad execution substrate.
	CreateJob(
		ctx context.Context,
		eventID string,
		jobName string,
		jobSpec brignext.JobSpec,
	) error
	// GetJobStatus, given an Event identifier and Job name, returns the Job's
	// status.
	GetJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
	) (brignext.JobStatus, error)
	// WatchJobStatus, given an Event identifier and Job name, returns a channel
	// over which the Job's status is streamed. The channel receives a new
	// JobStatus every time there is any change in that status.
	WatchJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan brignext.JobStatus, error)
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
		selector brignext.LogsSelector,
		opts meta.ListOptions,
	) (brignext.LogEntryList, error)
	// StreamLogs returns a channel over which logs for an Event's Worker, or
	// using the LogOptions parameter, a Job spawned by that Worker (or specific
	// container thereof), are streamed.
	StreamLogs(
		ctx context.Context,
		eventID string,
		selector brignext.LogsSelector,
	) (<-chan brignext.LogEntry, error)
}

type service struct {
	projectsStore projects.Store
	store         Store
	warmLogsStore LogsStore
	coolLogsStore LogsStore
	scheduler     Scheduler
}

// NewService returns a specialized interface for managing Events.
func NewService(
	projectsStore projects.Store,
	store Store,
	warmLogsStore LogsStore,
	coolLogsStore LogsStore,
	scheduler Scheduler,
) Service {
	return &service{
		projectsStore: projectsStore,
		store:         store,
		scheduler:     scheduler,
		warmLogsStore: warmLogsStore,
		coolLogsStore: coolLogsStore,
	}
}

// TODO: There's a lot of stuff that happens in this function that maybe we
// should defer until later-- like when the worker pod actually gets created.
func (s *service) Create(
	ctx context.Context,
	event brignext.Event,
) (brignext.EventList, error) {
	events := brignext.EventList{}

	now := time.Now()
	event.Created = &now

	// If no project ID is specified, we use other criteria to locate projects
	// that are subscribed to this event. We iterate over all of those and create
	// an event for each of these by making a recursive call to this same
	// function.
	if event.ProjectID == "" {
		projects, err := s.projectsStore.ListSubscribers(ctx, event)
		if err != nil {
			return events, errors.Wrap(
				err,
				"error retrieving subscribed projects from store",
			)
		}
		events.Items = make([]brignext.Event, len(projects.Items))
		for i, project := range projects.Items {
			event.ProjectID = project.ID
			projectEvents, err := s.Create(ctx, event)
			if err != nil {
				return events, err
			}
			// projectEvents.Items will always contain precisely one element
			events.Items[i] = projectEvents.Items[0]
		}
		return events, nil
	}

	// Make sure the project exists
	project, err := s.projectsStore.Get(ctx, event.ProjectID)
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

	token := crypto.NewToken(256)

	event.Worker = brignext.Worker{
		Spec: workerSpec,
		Status: brignext.WorkerStatus{
			Phase: brignext.WorkerPhasePending,
		},
		Token:       token,
		HashedToken: crypto.ShortSHA("", token),
	}

	// Let the scheduler add scheduler-specific details before we persist.
	if event, err = s.scheduler.PreCreate(ctx, project, event); err != nil {
		return events, errors.Wrapf(
			err,
			"error pre-creating event %q in scheduler",
			event.ID,
		)
	}

	// TODO: We'd like to use transaction semantics here, but transactions in
	// MongoDB are dicey, so we should refine this strategy to where a
	// partially completed create leaves us, overall, in a tolerable state.

	if err = s.store.Create(ctx, event); err != nil {
		return events, errors.Wrapf(
			err,
			"error storing new event %q",
			event.ID,
		)
	}
	if err = s.scheduler.Create(ctx, project, event); err != nil {
		return events, errors.Wrapf(
			err,
			"error creating event %q in scheduler",
			event.ID,
		)
	}

	events.Items = []brignext.Event{event}
	return events, nil
}

func (s *service) List(
	ctx context.Context,
	selector brignext.EventsSelector,
	opts meta.ListOptions,
) (brignext.EventList, error) {
	events := brignext.EventList{}

	// If no worker phase filters were applied, retrieve all phases
	if len(selector.WorkerPhases) == 0 {
		selector.WorkerPhases = brignext.WorkerPhasesAll()
	}
	if opts.Limit == 0 {
		opts.Limit = 20
	}

	events, err := s.store.List(ctx, selector, opts)
	if err != nil {
		return events, errors.Wrap(err, "error retrieving events from store")
	}
	return events, nil
}

func (s *service) Get(
	ctx context.Context,
	id string,
) (brignext.Event, error) {
	event, err := s.store.Get(ctx, id)
	if err != nil {
		return event, errors.Wrapf(err, "error retrieving event %q from store", id)
	}
	return event, nil
}

func (s *service) GetByWorkerToken(
	ctx context.Context,
	workerToken string,
) (brignext.Event, error) {
	event, err := s.store.GetByHashedWorkerToken(
		ctx,
		crypto.ShortSHA("", workerToken),
	)
	if err != nil {
		return event, errors.Wrap(err, "error retrieving event from store")
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

func (s *service) CancelMany(
	ctx context.Context,
	selector brignext.EventsSelector,
) (brignext.CancelManyEventsResult, error) {
	result := brignext.CancelManyEventsResult{}

	// Refuse requests not qualified by project
	if selector.ProjectID == "" {
		return result, &brignext.ErrBadRequest{
			Reason: "Requests to cancel multiple events must be qualified by " +
				"project.",
		}
	}
	// Refuse requests not qualified by worker phases
	if len(selector.WorkerPhases) == 0 {
		return result, &brignext.ErrBadRequest{
			Reason: "Requests to cancel multiple events must be qualified by " +
				"worker phase(s).",
		}
	}

	if selector.ProjectID != "" {
		// Make sure the project exists
		_, err := s.projectsStore.Get(ctx, selector.ProjectID)
		if err != nil {
			return result, errors.Wrapf(
				err,
				"error retrieving project %q from store",
				selector.ProjectID,
			)
		}
	}

	events, err := s.store.CancelMany(ctx, selector)
	if err != nil {
		return result, errors.Wrap(err, "error canceling events in store")
	}

	result.Count = int64(len(events.Items))

	// TODO: Can we find a quicker, more efficient way to do this?
	go func() {
		for _, event := range events.Items {
			if err := s.scheduler.Delete(
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

func (s *service) DeleteMany(
	ctx context.Context,
	selector brignext.EventsSelector,
) (brignext.DeleteManyEventsResult, error) {
	result := brignext.DeleteManyEventsResult{}

	// Refuse requests not qualified by project
	if selector.ProjectID == "" {
		return result, &brignext.ErrBadRequest{
			Reason: "Requests to delete multiple events must be qualified by " +
				"project.",
		}
	}
	// Refuse requests not qualified by worker phases
	if len(selector.WorkerPhases) == 0 {
		return result, &brignext.ErrBadRequest{
			Reason: "Requests to delete multiple events must be qualified by " +
				"worker phase(s).",
		}
	}

	if selector.ProjectID != "" {
		// Make sure the project exists
		_, err := s.projectsStore.Get(ctx, selector.ProjectID)
		if err != nil {
			return result, errors.Wrapf(
				err,
				"error retrieving project %q from store",
				selector.ProjectID,
			)
		}
	}

	events, err := s.store.DeleteMany(ctx, selector)
	if err != nil {
		return result, errors.Wrap(err, "error deleting events from store")
	}

	result.Count = int64(len(events.Items))

	// TODO: Can we find a quicker, more efficient way to do this?
	go func() {
		for _, event := range events.Items {
			if err := s.scheduler.Delete(
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

func (s *service) StartWorker(ctx context.Context, eventID string) error {
	event, err := s.store.Get(ctx, eventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}

	// spec := event.Worker.Spec
	// // TODO: This is probably a better place to apply worker default just before
	// // it is started INSTEAD OF setting defaults at event creation time or waiting
	// // all the way up until pod creation time.
	// if err = s.store.UpdateWorkerSpec(ctx, eventID, spec); err != nil {
	// 	return errors.Wrapf(
	// 		err,
	// 		"error updating worker's spec for event %q",
	// 		event.ID,
	// 	)
	// }

	if err = s.scheduler.StartWorker(ctx, event); err != nil {
		return errors.Wrapf(err, "error starting worker for event %q", event.ID)
	}
	return nil
}

func (s *service) GetWorkerStatus(
	ctx context.Context,
	eventID string,
) (brignext.WorkerStatus, error) {
	event, err := s.store.Get(ctx, eventID)
	if err != nil {
		return brignext.WorkerStatus{},
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	return event.Worker.Status, nil
}

// TODO: Should we put some kind of timeout on this function?
func (s *service) WatchWorkerStatus(
	ctx context.Context,
	eventID string,
) (<-chan brignext.WorkerStatus, error) {
	// Read the event up front to confirm it exists.
	if _, err := s.store.Get(ctx, eventID); err != nil {
		return nil,
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	statusCh := make(chan brignext.WorkerStatus)
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
			event, err := s.store.Get(ctx, eventID)
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

func (s *service) CreateJob(
	ctx context.Context,
	eventID string,
	jobName string,
	jobSpec brignext.JobSpec,
) error {
	event, err := s.store.Get(ctx, eventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	if _, ok := event.Worker.Jobs[jobName]; ok {
		return &brignext.ErrConflict{
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
		return &brignext.ErrAuthorization{} // TODO: Add more detail to this error
	}
	if useDockerSocket &&
		(event.Worker.Spec.JobPolicies == nil ||
			!event.Worker.Spec.JobPolicies.AllowDockerSocketMount) {
		return &brignext.ErrAuthorization{} // TODO: Add more detail to this error
	}

	// Fail quickly if the job needs to use shared workspace, but the worker
	// doesn't have any shared workspace.
	if useWorkspace && !event.Worker.Spec.UseWorkspace {
		return &brignext.ErrConflict{} // TODO: Add more detail to this error
	}

	if err = s.store.CreateJob(ctx, eventID, jobName, jobSpec); err != nil {
		return errors.Wrapf(
			err, "error saving event %q job %q in store",
			eventID,
			eventID,
		)
	}

	if event.Worker.Jobs == nil {
		event.Worker.Jobs = map[string]brignext.Job{}
	}
	event.Worker.Jobs[jobName] = brignext.Job{
		Spec: jobSpec,
	}

	if err = s.scheduler.StartJob(ctx, event, jobName); err != nil {
		return errors.Wrapf(
			err,
			"error starting event %q job %q",
			event.ID,
			jobName,
		)
	}

	return nil
}

func (s *service) GetJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (brignext.JobStatus, error) {
	event, err := s.store.Get(ctx, eventID)
	if err != nil {
		return brignext.JobStatus{},
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	job, ok := event.Worker.Jobs[jobName]
	if !ok {
		return brignext.JobStatus{}, &brignext.ErrNotFound{
			Type: "Job",
			ID:   jobName,
		}
	}
	return job.Status, nil
}

// TODO: Should we put some kind of timeout on this function?
func (s *service) WatchJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan brignext.JobStatus, error) {
	// Read the event and job up front to confirm they both exists.
	event, err := s.store.Get(ctx, eventID)
	if err != nil {
		return nil,
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	if _, ok := event.Worker.Jobs[jobName]; !ok {
		return nil, &brignext.ErrNotFound{
			Type: "Job",
			ID:   jobName,
		}
	}
	statusCh := make(chan brignext.JobStatus)
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
			event, err := s.store.Get(ctx, eventID)
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
	selector brignext.LogsSelector,
	opts meta.ListOptions,
) (brignext.LogEntryList, error) {
	logEntries := brignext.LogEntryList{}
	if opts.Limit == 0 {
		opts.Limit = 1000
	}
	event, err := s.store.Get(ctx, eventID)
	if err != nil {
		return logEntries,
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	// Try warm logs first and fall back on cooler logs if necessary.
	logEntries, err = s.warmLogsStore.GetLogs(ctx, event, selector, opts)
	if err != nil {
		log.Println(err)
		logEntries, err = s.coolLogsStore.GetLogs(ctx, event, selector, opts)
	}
	return logEntries, err
}

func (s *service) StreamLogs(
	ctx context.Context,
	eventID string,
	selector brignext.LogsSelector,
) (<-chan brignext.LogEntry, error) {
	event, err := s.store.Get(ctx, eventID)
	if err != nil {
		return nil,
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	// Try warm logs first and fall back on cooler logs if necessary
	logCh, err := s.warmLogsStore.StreamLogs(ctx, event, selector)
	if err != nil {
		logCh, err = s.coolLogsStore.StreamLogs(ctx, event, selector)
	}
	return logCh, err
}
