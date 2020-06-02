package storage

import (
	"context"

	"github.com/krancour/brignext/v2"
)

type EventsStore interface {
	Create(context.Context, brignext.Event) error
	List(context.Context) (brignext.EventList, error)
	ListByProject(context.Context, string) (brignext.EventList, error)
	Get(context.Context, string) (brignext.Event, error)
	Cancel(
		ctx context.Context,
		id string,
	) error
	CancelCollection(
		ctx context.Context,
		opts brignext.EventListOptions,
	) (brignext.EventReferenceList, error)
	Delete(
		ctx context.Context,
		id string,
		deletePending bool,
		deleteRunning bool,
	) (bool, error)
	DeleteByProject(context.Context, string) error
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
