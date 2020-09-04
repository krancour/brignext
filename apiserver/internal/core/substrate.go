package core

import "context"

type Substrate interface {
	// PreCreateProject returns a Project that has been amended with
	// substrate-specific details. This should always be called prior to a Project
	// being initially persisted so that substrate-specific details will be
	// included.
	PreCreateProject(ctx context.Context, project Project) (Project, error)
	// CreateProject prepares the substrate to host Project workloads.
	CreateProject(ctx context.Context, project Project) error
	// PreUpdateProject returns a Project that has been amended with
	// substrate-specific details. This should always be called prior to an
	// updated Project being persisted so that substrate-specific details are not
	// lost during the update process.
	PreUpdateProject(
		ctx context.Context,
		oldProject Project,
		newProject Project,
	) (Project, error)
	// UpdateProject makes any necessary adjustments to the substrate to reflect
	// updates to a Project.
	UpdateProject(
		ctx context.Context,
		oldProject Project,
		newProject Project,
	) error
	// DeleteProject removes all Project-related resources from the substrate.
	DeleteProject(ctx context.Context, project Project) error

	// PreCreateEvent returns an Event that has been amended with
	// substrate-specific details. This should always be called prior to an Event
	// being initially persisted so that substrate-specific details will be
	// included.
	PreCreateEvent(
		ctx context.Context,
		project Project,
		event Event,
	) (Event, error)

	// PreScheduleWorker prepares the substrate to execute an Event's Worker, but
	// does not start the Worker.
	PreScheduleWorker(ctx context.Context, event Event) error
	// ScheduleWorker schedules an Event's Worker for async / eventual execution
	// on the substrate.
	ScheduleWorker(ctx context.Context, event Event) error
	// StartWorker starts an Event's Worker on the substrate.
	StartWorker(ctx context.Context, event Event) error

	// PreScheduleJob prepares the substrate to execute a Job, but does not start
	// the Job.
	PreScheduleJob(ctx context.Context, event Event, jobName string) error
	// ScheduleJob schedules a Job for async / eventual execution on the
	// substrate.
	ScheduleJob(ctx context.Context, event Event, jobName string) error
	// StartJob starts a Job on the substrate.
	StartJob(ctx context.Context, event Event, jobName string) error

	// DeleteWorkerAndJobs deletes all substrate resources pertaining to the
	// specified Event's Worker and Jobs.
	DeleteWorkerAndJobs(context.Context, Event) error
}
