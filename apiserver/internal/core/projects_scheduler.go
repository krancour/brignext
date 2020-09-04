package core

import (
	"context"
)

type ProjectsScheduler interface {
	PreCreate(
		ctx context.Context,
		project Project,
	) (Project, error)
	Create(
		ctx context.Context,
		project Project,
	) error
	PreUpdate(
		ctx context.Context,
		project Project,
	) (Project, error)
	Update(
		ctx context.Context,
		project Project,
	) error
	Delete(
		ctx context.Context,
		project Project,
	) error
}
