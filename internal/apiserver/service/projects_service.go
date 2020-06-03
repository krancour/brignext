package service

import (
	"context"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/scheduler"
	"github.com/krancour/brignext/v2/internal/apiserver/storage"
	"github.com/pkg/errors"
)

type ProjectsService interface {
	Create(context.Context, brignext.Project) error
	List(context.Context) (brignext.ProjectList, error)
	Get(context.Context, string) (brignext.Project, error)
	Update(context.Context, brignext.Project) error
	Delete(context.Context, string) error

	ListSecrets(
		ctx context.Context,
		projectID string,
	) (brignext.SecretList, error)
	SetSecret(
		ctx context.Context,
		projectID string,
		secret brignext.Secret,
	) error
	UnsetSecret(ctx context.Context, projectID string, key string) error
}

type projectsService struct {
	store     storage.Store
	scheduler scheduler.ProjectsScheduler
}

func NewProjectsService(
	store storage.Store,
	scheduler scheduler.ProjectsScheduler,
) ProjectsService {
	return &projectsService{
		store:     store,
		scheduler: scheduler,
	}
}

func (p *projectsService) Create(
	ctx context.Context,
	project brignext.Project,
) error {
	project = projectWithDefaults(project)

	// We send this to the scheduler first because we expect the scheduler will
	// will add some scheduler-specific details that we will want to persist.
	var err error
	project, err = p.scheduler.Create(ctx, project)
	if err != nil {
		return errors.Wrapf(
			err,
			"error creating project %q in the scheduler",
			project.ID,
		)
	}
	if err := p.store.Projects().Create(ctx, project); err != nil {
		// We need to roll this back manually because the scheduler doesn't
		// automatically roll anything back upon failure.
		p.scheduler.Delete(ctx, project) // nolint: errcheck
		return errors.Wrapf(err, "error storing new project %q", project.ID)
	}
	return nil
}

func (p *projectsService) List(
	ctx context.Context,
) (brignext.ProjectList, error) {
	projectList, err := p.store.Projects().List(ctx)
	if err != nil {
		return projectList, errors.Wrap(err, "error retrieving projects from store")
	}
	return projectList, nil
}

func (p *projectsService) Get(
	ctx context.Context,
	id string,
) (brignext.Project, error) {
	project, err := p.store.Projects().Get(ctx, id)
	if err != nil {
		return project, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			id,
		)
	}
	return project, nil
}

func (p *projectsService) Update(
	ctx context.Context,
	project brignext.Project,
) error {
	project = projectWithDefaults(project)

	// We send this to the scheduler first because we expect the scheduler will
	// will add some scheduler-specific details that we will want to persist.
	var err error
	project, err = p.scheduler.Update(ctx, project)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating project %q in the scheduler",
			project.ID,
		)
	}

	if err := p.store.Projects().Update(ctx, project); err != nil {
		return errors.Wrapf(
			err,
			"error updating project %q in store",
			project.ID,
		)
	}
	return nil
}

func (p *projectsService) Delete(ctx context.Context, id string) error {
	project, err := p.store.Projects().Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving project %q from store", id)
	}
	return p.store.DoTx(ctx, func(ctx context.Context) error {
		if err := p.store.Projects().Delete(ctx, id); err != nil {
			return errors.Wrapf(err, "error removing project %q from store", id)
		}
		if _, err := p.store.Events().DeleteCollection(
			ctx,
			brignext.EventListOptions{
				ProjectID:    id,
				WorkerPhases: brignext.WorkerPhasesAll(),
			},
		); err != nil {
			return errors.Wrapf(
				err,
				"error deleting events for project %q from scheduler",
				id,
			)
		}
		if err := p.scheduler.Delete(ctx, project); err != nil {
			return errors.Wrapf(
				err,
				"error deleting project %q from scheduler",
				id,
			)
		}
		return nil
	})
}

func (p *projectsService) ListSecrets(
	ctx context.Context,
	projectID string,
) (brignext.SecretList, error) {
	secretList := brignext.SecretList{}
	project, err := p.store.Projects().Get(ctx, projectID)
	if err != nil {
		return secretList, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	if secretList, err =
		p.scheduler.ListSecrets(ctx, project); err != nil {
		return secretList, errors.Wrapf(
			err,
			"error getting worker secrets for project %q from scheduler",
			projectID,
		)
	}
	return secretList, nil
}

func (p *projectsService) SetSecret(
	ctx context.Context,
	projectID string,
	secret brignext.Secret,
) error {
	project, err := p.store.Projects().Get(ctx, projectID)
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

func (p *projectsService) UnsetSecret(
	ctx context.Context,
	projectID string,
	key string,
) error {
	project, err := p.store.Projects().Get(ctx, projectID)
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
		p.scheduler.UnsetSecret(ctx, project, key); err != nil {
		return errors.Wrapf(
			err,
			"error unsetting secrets for project %q worker in scheduler",
			projectID,
		)
	}
	return nil
}

func projectWithDefaults(project brignext.Project) brignext.Project {
	if project.Spec.EventSubscriptions == nil {
		project.Spec.EventSubscriptions = []brignext.EventSubscription{}
	}

	if project.Spec.Worker.Container.Environment == nil {
		project.Spec.Worker.Container.Environment = map[string]string{}
	}

	if project.Spec.Worker.DefaultConfigFiles == nil {
		project.Spec.Worker.DefaultConfigFiles = map[string]string{}
	}

	return project
}
