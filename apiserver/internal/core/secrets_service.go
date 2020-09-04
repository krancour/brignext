package core

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

type SecretsService interface {
	// List returns a SecretList who Items (Secrets) contain Keys only and
	// not Values (all Value fields are empty). i.e. Once a secret is set, end
	// clients are unable to retrieve values.
	List(
		ctx context.Context,
		projectID string,
		opts meta.ListOptions,
	) (SecretList, error)
	// Set set the value of a new Secret or updates the value of an existing
	// Secret.
	Set(
		ctx context.Context,
		projectID string,
		secret Secret,
	) error
	// Unset clears the value of an existing Secret.
	Unset(ctx context.Context, projectID string, key string) error
}

type secretsService struct {
	authorize     authx.AuthorizeFn
	projectsStore ProjectsStore
	scheduler     ProjectsScheduler
}

func NewSecretsService(
	projectsStore ProjectsStore,
	scheduler ProjectsScheduler,
) SecretsService {
	return &secretsService{
		authorize:     authx.Authorize,
		projectsStore: projectsStore,
		scheduler:     scheduler,
	}
}

func (s *secretsService) List(
	ctx context.Context,
	projectID string,
	opts meta.ListOptions,
) (SecretList, error) {
	if err := s.authorize(ctx, authx.RoleReader()); err != nil {
		return SecretList{}, err
	}

	secrets := SecretList{}
	project, err := s.projectsStore.Get(ctx, projectID)
	if err != nil {
		return secrets, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	if opts.Limit == 0 {
		opts.Limit = 20
	}
	if secrets, err =
		s.scheduler.ListSecrets(ctx, project, opts); err != nil {
		return secrets, errors.Wrapf(
			err,
			"error getting worker secrets for project %q from scheduler",
			projectID,
		)
	}
	return secrets, nil
}

func (p *secretsService) Set(
	ctx context.Context,
	projectID string,
	secret Secret,
) error {
	if err := p.authorize(ctx, authx.RoleProjectAdmin(projectID)); err != nil {
		return err
	}

	project, err := p.projectsStore.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	// Secrets aren't stored in the database. We only pass them to the scheduler.
	if err := p.scheduler.SetSecret(ctx, project, secret); err != nil {
		return errors.Wrapf(
			err,
			"error setting secret for project %q worker in scheduler",
			projectID,
		)
	}
	return nil
}

func (s *secretsService) Unset(
	ctx context.Context,
	projectID string,
	key string,
) error {
	if err := s.authorize(ctx, authx.RoleProjectAdmin(projectID)); err != nil {
		return err
	}

	project, err := s.projectsStore.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	// Secrets aren't stored in the database. We only have to remove them from the
	// scheduler.
	if err :=
		s.scheduler.UnsetSecret(ctx, project, key); err != nil {
		return errors.Wrapf(
			err,
			"error unsetting secrets for project %q worker in scheduler",
			projectID,
		)
	}
	return nil
}
