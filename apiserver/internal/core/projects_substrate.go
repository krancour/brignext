package core

import (
	"context"
)

type ProjectsSubstrate interface {
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
		oldProject Project,
		newProject Project,
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
