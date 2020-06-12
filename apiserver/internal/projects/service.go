package projects

import (
	"context"

	brignext "github.com/krancour/brignext/v2/sdk"
	"github.com/pkg/errors"
)

type Service interface {
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

	CheckHealth(context.Context) error
}

type service struct {
	store     Store
	scheduler Scheduler
}

func NewService(store Store, scheduler Scheduler) Service {
	return &service{
		store:     store,
		scheduler: scheduler,
	}
}

func (s *service) Create(
	ctx context.Context,
	project brignext.Project,
) error {
	project = s.projectWithDefaults(project)

	// We send this to the scheduler first because we expect the scheduler will
	// will add some scheduler-specific details that we will want to persist.
	var err error
	project, err = s.scheduler.Create(ctx, project)
	if err != nil {
		return errors.Wrapf(
			err,
			"error creating project %q in the scheduler",
			project.ID,
		)
	}
	if err := s.store.Create(ctx, project); err != nil {
		// We need to roll this back manually because the scheduler doesn't
		// automatically roll anything back upon failure.
		s.scheduler.Delete(ctx, project) // nolint: errcheck
		return errors.Wrapf(err, "error storing new project %q", project.ID)
	}
	return nil
}

func (s *service) List(
	ctx context.Context,
) (brignext.ProjectList, error) {
	projectList, err := s.store.List(ctx)
	if err != nil {
		return projectList, errors.Wrap(err, "error retrieving projects from store")
	}
	return projectList, nil
}

func (s *service) Get(
	ctx context.Context,
	id string,
) (brignext.Project, error) {
	project, err := s.store.Get(ctx, id)
	if err != nil {
		return project, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			id,
		)
	}
	return project, nil
}

func (s *service) Update(
	ctx context.Context,
	project brignext.Project,
) error {
	project = s.projectWithDefaults(project)

	// We send this to the scheduler first because we expect the scheduler will
	// will add some scheduler-specific details that we will want to persist.
	var err error
	project, err = s.scheduler.Update(ctx, project)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating project %q in the scheduler",
			project.ID,
		)
	}

	if err := s.store.Update(ctx, project); err != nil {
		return errors.Wrapf(
			err,
			"error updating project %q in store",
			project.ID,
		)
	}
	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	project, err := s.store.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving project %q from store", id)
	}
	return s.store.DoTx(ctx, func(ctx context.Context) error {
		if err := s.store.Delete(ctx, id); err != nil {
			return errors.Wrapf(err, "error removing project %q from store", id)
		}
		// TODO: Cascade the delete to the project's events somehow
		// if _, err := p.store.Events().DeleteCollection(
		// 	ctx,
		// 	brignext.EventListOptions{
		// 		ProjectID:    id,
		// 		WorkerPhases: brignext.WorkerPhasesAll(),
		// 	},
		// ); err != nil {
		// 	return errors.Wrapf(
		// 		err,
		// 		"error deleting events for project %q from scheduler",
		// 		id,
		// 	)
		// }
		if err := s.scheduler.Delete(ctx, project); err != nil {
			return errors.Wrapf(
				err,
				"error deleting project %q from scheduler",
				id,
			)
		}
		return nil
	})
}

func (s *service) ListSecrets(
	ctx context.Context,
	projectID string,
) (brignext.SecretList, error) {
	secretList := brignext.SecretList{}
	project, err := s.store.Get(ctx, projectID)
	if err != nil {
		return secretList, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	if secretList, err =
		s.scheduler.ListSecrets(ctx, project); err != nil {
		return secretList, errors.Wrapf(
			err,
			"error getting worker secrets for project %q from scheduler",
			projectID,
		)
	}
	return secretList, nil
}

func (s *service) SetSecret(
	ctx context.Context,
	projectID string,
	secret brignext.Secret,
) error {
	project, err := s.store.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	// Secrets aren't stored in the database. We only pass them to the scheduler.
	if err := s.scheduler.SetSecret(ctx, project, secret); err != nil {
		return errors.Wrapf(
			err,
			"error setting secret for project %q worker in scheduler",
			projectID,
		)
	}
	return nil
}

func (s *service) UnsetSecret(
	ctx context.Context,
	projectID string,
	key string,
) error {
	project, err := s.store.Get(ctx, projectID)
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

func (s *service) CheckHealth(ctx context.Context) error {
	if err := s.store.CheckHealth(ctx); err != nil {
		return err
	}
	return s.scheduler.CheckHealth(ctx)
}

func (s *service) projectWithDefaults(
	project brignext.Project,
) brignext.Project {
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
