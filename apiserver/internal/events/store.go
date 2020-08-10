package events

import (
	"context"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
)

type Store interface {
	Create(context.Context, brignext.Event) error
	List(
		context.Context,
		brignext.EventListOptions,
	) (brignext.EventList, error)
	Get(context.Context, string) (brignext.Event, error)
	Cancel(context.Context, string) error
	CancelMany(
		context.Context,
		brignext.EventListOptions,
	) (brignext.EventList, error)
	Delete(context.Context, string) error
	DeleteMany(
		context.Context,
		brignext.EventListOptions,
	) (brignext.EventList, error)

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
}
