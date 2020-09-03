package core

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/meta"
)

type ProjectsStore interface {
	Create(context.Context, Project) error
	List(
		context.Context,
		ProjectsSelector,
		meta.ListOptions,
	) (ProjectList, error)
	ListSubscribers(
		ctx context.Context,
		event Event,
	) (ProjectList, error)
	Get(context.Context, string) (Project, error)
	Update(context.Context, Project) error
	Delete(context.Context, string) error
}
