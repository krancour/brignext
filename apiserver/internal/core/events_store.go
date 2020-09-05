package core

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
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
}
