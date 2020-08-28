package events

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
)

type Store interface {
	Create(context.Context, core.Event) error
	List(
		context.Context,
		core.EventsSelector,
		meta.ListOptions,
	) (core.EventList, error)
	Get(context.Context, string) (core.Event, error)
	GetByHashedWorkerToken(context.Context, string) (core.Event, error)
	Cancel(context.Context, string) error
	CancelMany(
		context.Context,
		core.EventsSelector,
	) (core.EventList, error)
	Delete(context.Context, string) error
	DeleteMany(
		context.Context,
		core.EventsSelector,
	) (core.EventList, error)

	UpdateWorkerSpec(
		ctx context.Context,
		eventID string,
		spec core.WorkerSpec,
	) error
	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		status core.WorkerStatus,
	) error

	CreateJob(
		ctx context.Context,
		eventID string,
		jobName string,
		jobSpec core.JobSpec,
	) error
	UpdateJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status core.JobStatus,
	) error
}
