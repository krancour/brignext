package core

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

type JobsService interface {
	// Create, given an Event identifier and JobSpec, creates a new Job and
	// starts it on BrigNext's workload execution substrate.
	Create(
		ctx context.Context,
		eventID string,
		jobName string,
		jobSpec JobSpec,
	) error
	// Start, given an Event identifier and Job name, starts that Job on
	// BrigNext's workload execution substrate.
	Start(
		ctx context.Context,
		eventID string,
		jobName string,
	) error
	// GetStatus, given an Event identifier and Job name, returns the Job's
	// status.
	GetStatus(
		ctx context.Context,
		eventID string,
		jobName string,
	) (JobStatus, error)
	// WatchStatus, given an Event identifier and Job name, returns a channel
	// over which the Job's status is streamed. The channel receives a new
	// JobStatus every time there is any change in that status.
	WatchStatus(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan JobStatus, error)
	// UpdateStatus, given an Event identifier and Job name, updates the status
	// of that Job.
	UpdateStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status JobStatus,
	) error
}

type jobsService struct {
	authorize authx.AuthorizeFn
	// projectsStore ProjectsStore
	eventsStore EventsStore
	scheduler   EventsScheduler
}

func NewJobsService(
	// projectsStore ProjectsStore,
	eventsStore EventsStore,
	scheduler EventsScheduler,
) JobsService {
	return &jobsService{
		authorize: authx.Authorize,
		// projectsStore: projectsStore,
		eventsStore: eventsStore,
		scheduler:   scheduler,
	}
}

func (j *jobsService) Create(
	ctx context.Context,
	eventID string,
	jobName string,
	jobSpec JobSpec,
) error {
	if err := j.authorize(ctx, authx.RoleWorker(eventID)); err != nil {
		return err
	}

	event, err := j.eventsStore.Get(ctx, eventID)
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

	if err = j.eventsStore.CreateJob(ctx, eventID, jobName, jobSpec); err != nil {
		return errors.Wrapf(
			err, "error saving event %q job %q in store",
			eventID,
			eventID,
		)
	}

	if err = j.scheduler.CreateJob(ctx, event, jobName); err != nil {
		return errors.Wrapf(
			err,
			"error scheduling event %q job %q",
			event.ID,
			jobName,
		)
	}

	return nil
}

func (j *jobsService) Start(
	ctx context.Context,
	eventID string,
	jobName string,
) error {
	if err := j.authorize(ctx, authx.RoleScheduler()); err != nil {
		return err
	}

	event, err := j.eventsStore.Get(ctx, eventID)
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

	if err = j.scheduler.StartJob(ctx, event, jobName); err != nil {
		return errors.Wrapf(
			err,
			"error starting event %q job %q",
			event.ID,
			jobName,
		)
	}

	return nil
}

func (j *jobsService) GetStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (JobStatus, error) {
	if err := j.authorize(ctx, authx.RoleReader()); err != nil {
		return JobStatus{}, err
	}

	event, err := j.eventsStore.Get(ctx, eventID)
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
func (j *jobsService) WatchStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan JobStatus, error) {
	if err := j.authorize(ctx, authx.RoleReader()); err != nil {
		return nil, err
	}

	// Read the event and job up front to confirm they both exists.
	event, err := j.eventsStore.Get(ctx, eventID)
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
			event, err := j.eventsStore.Get(ctx, eventID)
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

func (j *jobsService) UpdateStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status JobStatus,
) error {
	if err := j.authorize(ctx, authx.RoleObserver()); err != nil {
		return err
	}

	if err := j.eventsStore.UpdateJobStatus(
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
