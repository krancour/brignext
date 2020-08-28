package projects

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
)

type Store interface {
	Create(context.Context, brignext.Project) error
	List(
		context.Context,
		brignext.ProjectsSelector,
		meta.ListOptions,
	) (brignext.ProjectList, error)
	ListSubscribers(
		ctx context.Context,
		event brignext.Event,
	) (brignext.ProjectList, error)
	Get(context.Context, string) (brignext.Project, error)
	Update(context.Context, brignext.Project) error
	Delete(context.Context, string) error
}
