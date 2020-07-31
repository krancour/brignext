package projects

import (
	"context"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
)

type Store interface {
	Create(context.Context, brignext.Project) error
	List(context.Context) (brignext.ProjectReferenceList, error)
	ListSubscribers(
		ctx context.Context,
		event brignext.Event,
	) (brignext.ProjectReferenceList, error)
	Get(context.Context, string) (brignext.Project, error)
	Update(context.Context, brignext.Project) error
	Delete(context.Context, string) error
}
