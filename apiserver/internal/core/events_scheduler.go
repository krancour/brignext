package core

import (
	"context"
)

type EventsScheduler interface {
	PreCreate(
		ctx context.Context,
		project Project,
		event Event,
	) (Event, error)
	Create(
		ctx context.Context,
		project Project,
		event Event,
	) error
	Delete(context.Context, Event) error

	StartWorker(ctx context.Context, event Event) error

	CreateJob(
		ctx context.Context,
		event Event,
		jobName string,
	) error
	StartJob(
		ctx context.Context,
		event Event,
		jobName string,
	) error
}
