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
	// DeleteProject removes all Project-related resources from the substrate.
	DeleteProject(context.Context, Project) error

	// PreCreateEvent returns an Event that has been amended with
	// substrate-specific details. This should always be called prior to an Event
	// being initially persisted so that substrate-specific details will be
	// included.
	PreCreateEvent(context.Context, Project, Event) (Event, error)

	// ScheduleWorker prepares the substrate for the Event's worker and schedules
	// the Worker for async / eventual execution.
	ScheduleWorker(context.Context, Project, Event) error
	// StartWorker starts an Event's Worker on the substrate.
	StartWorker(context.Context, Project, Event) error

	// ScheduleJob prepares the substrate for a Job and schedules the Job for
	// async / eventual execution.
	ScheduleJob(
		ctx context.Context,
		project Project,
		event Event,
		jobName string,
	) error
	// StartJob starts a Job on the substrate.
	StartJob(
		ctx context.Context,
		project Project,
		event Event,
		jobName string,
	) error

	// DeleteWorkerAndJobs deletes all substrate resources pertaining to the
	// specified Event's Worker and Jobs.
	DeleteWorkerAndJobs(context.Context, Project, Event) error
}
