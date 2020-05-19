package storage

import (
	"context"

	"github.com/krancour/brignext/v2"
)

type ProjectsStore interface {
	Create(context.Context, brignext.Project) error
	List(context.Context) (brignext.ProjectList, error)
	ListSubscribed(
		ctx context.Context,
		event brignext.Event,
	) (brignext.ProjectList, error)
	Get(context.Context, string) (brignext.Project, error)
	Update(context.Context, brignext.Project) error
	Delete(context.Context, string) error
}
