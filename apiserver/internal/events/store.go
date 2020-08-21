package events

import (
	"context"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
)

type Store interface {
	Create(context.Context, brignext.Event) error
	List(
		context.Context,
		brignext.EventsSelector,
		meta.ListOptions,
	) (brignext.EventList, error)
	Get(context.Context, string) (brignext.Event, error)
	GetByHashedWorkerToken(context.Context, string) (brignext.Event, error)
	Cancel(context.Context, string) error
	CancelMany(
		context.Context,
		brignext.EventsSelector,
	) (brignext.EventList, error)
	Delete(context.Context, string) error
	DeleteMany(
		context.Context,
		brignext.EventsSelector,
	) (brignext.EventList, error)

	UpdateWorkerSpec(
		ctx context.Context,
		eventID string,
		spec brignext.WorkerSpec,
	) error
	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		status brignext.WorkerStatus,
	) error

	CreateJob(
		ctx context.Context,
		eventID string,
		jobName string,
		jobSpec brignext.JobSpec,
	) error
	UpdateJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status brignext.JobStatus,
	) error
}
