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
	Cancel(context.Context, string) error
	CancelCollection(
		context.Context,
		brignext.EventListOptions,
	) (brignext.EventReferenceList, error)
	Delete(context.Context, string) error
	DeleteCollection(
		context.Context,
		brignext.EventListOptions,
	) (brignext.EventReferenceList, error)
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
