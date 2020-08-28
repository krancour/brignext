package projects

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
)

type Store interface {
	Create(context.Context, core.Project) error
	List(
		context.Context,
		core.ProjectsSelector,
		meta.ListOptions,
	) (core.ProjectList, error)
	ListSubscribers(
		ctx context.Context,
		event core.Event,
	) (core.ProjectList, error)
	Get(context.Context, string) (core.Project, error)
	Update(context.Context, core.Project) error
	Delete(context.Context, string) error
}
