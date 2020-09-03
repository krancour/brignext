package core

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/meta"
)

type EventsStore interface {
	Create(context.Context, Event) error
	List(
		context.Context,
		EventsSelector,
		meta.ListOptions,
	) (EventList, error)
	Get(context.Context, string) (Event, error)
	GetByHashedWorkerToken(context.Context, string) (Event, error)
	Cancel(context.Context, string) error
	CancelMany(
		context.Context,
		EventsSelector,
	) (EventList, error)
	Delete(context.Context, string) error
	DeleteMany(
		context.Context,
		EventsSelector,
	) (EventList, error)

	UpdateWorkerSpec(
		ctx context.Context,
		eventID string,
		spec WorkerSpec,
	) error
	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		status WorkerStatus,
	) error

	CreateJob(
		ctx context.Context,
		eventID string,
		jobName string,
		jobSpec JobSpec,
	) error
	UpdateJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status JobStatus,
	) error
}
