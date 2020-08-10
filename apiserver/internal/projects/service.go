package projects

import (
	"context"
	"time"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/pkg/errors"
)

// Service is the specialized interface for managing Projects. It's decoupled
// from underlying technology choices (e.g. data store, message bus, etc.) to
// keep business logic reusable and consistent while the underlying tech stack
// remains free to change.
type Service interface {
	// Create creates a new Project.
	Create(context.Context, brignext.Project) (brignext.Project, error)
	// List returns a ProjectList, with its Items (Projects) ordered
	// alphabetically by Project ID.
	List(
		context.Context,
		brignext.ProjectListOptions,
	) (brignext.ProjectList, error)
	// Get retrieves a single Project specified by its identifier.
	Get(context.Context, string) (brignext.Project, error)
	// Update updates an existing Project.
	Update(context.Context, brignext.Project) (brignext.Project, error)
	// Delete deletes a single Project specified by its identifier.
	Delete(context.Context, string) error

	// ListSecrets returns a SecretList who Items (Secrets) contain Keys only and
	// not Values (all Value fields are empty). i.e. Once a secret is set, end
	// clients are unable to retrieve values.
	ListSecrets(
		ctx context.Context,
		projectID string,
	) (brignext.SecretList, error)
	// SetSecret set the value of a new Secret or updates the value of an existing
	// Secret.
	SetSecret(
		ctx context.Context,
		projectID string,
		secret brignext.Secret,
	) error
	// UnsetSecret clears the value of an existing Secret.
	UnsetSecret(ctx context.Context, projectID string, key string) error
}

type service struct {
	store     Store
	scheduler Scheduler
}

// NewService returns a specialized interface for managing Projects.
func NewService(store Store, scheduler Scheduler) Service {
	return &service{
		store:     store,
		scheduler: scheduler,
	}
}

func (s *service) Create(
	ctx context.Context,
	project brignext.Project,
) (brignext.Project, error) {
	now := time.Now()
	project.Created = &now

	project = s.projectWithDefaults(project)

	// Let the scheduler add scheduler-specific details before we persist.
	var err error
	if project, err = s.scheduler.PreCreate(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error pre-creating project %q in the scheduler",
			project.ID,
		)
	}

	// TODO: We'd like to use transaction semantics here, but transactions in
	// MongoDB are dicey, so we should refine this strategy to where a
	// partially completed create leaves us, overall, in a tolerable state.

	if err = s.store.Create(ctx, project); err != nil {
		return project, errors.Wrapf(err, "error storing new project %q", project.ID)
	}
	if err = s.scheduler.Create(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating project %q in the scheduler",
			project.ID,
		)
	}
	return project, nil
}

func (s *service) List(
	ctx context.Context,
	opts brignext.ProjectListOptions,
) (brignext.ProjectList, error) {
	projectList, err := s.store.List(ctx, opts)
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
) (brignext.Project, error) {
	project = s.projectWithDefaults(project)

	// Let the scheduler update scheduler-specific details before we persist.
	var err error
	if project, err = s.scheduler.PreUpdate(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error pre-updating project %q in the scheduler",
			project.ID,
		)
	}

	// TODO: We'd like to use transaction semantics here, but transactions in
	// MongoDB are dicey, so we should refine this strategy to where a
	// partially completed update leaves us, overall, in a tolerable state.

	if err = s.store.Update(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error updating project %q in store",
			project.ID,
		)
	}
	if err = s.scheduler.Update(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error updating project %q in the scheduler",
			project.ID,
		)
	}
	return project, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	project, err := s.store.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving project %q from store", id)
	}

	// TODO: We'd like to use transaction semantics here, but transactions in
	// MongoDB are dicey, so we should refine this strategy to where a
	// partially completed delete leaves us, overall, in a tolerable state.

	if err := s.store.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error removing project %q from store", id)
	}
	if err := s.scheduler.Delete(ctx, project); err != nil {
		return errors.Wrapf(
			err,
			"error deleting project %q from scheduler",
			id,
		)
	}
	return nil
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

// TODO: Should some of this be deferred until later, like when we actually
// start the worker?
func (s *service) projectWithDefaults(
	project brignext.Project,
) brignext.Project {
	// TODO: Do we need this?
	// if project.Spec.EventSubscriptions == nil {
	// 	project.Spec.EventSubscriptions = []brignext.EventSubscription{}
	// }

	// TODO: Do we need this?
	// if project.Spec.WorkerTemplate.Container.Environment == nil {
	// 	project.Spec.WorkerTemplate.Container.Environment = map[string]string{}
	// }

	// TODO: Do we need this?
	// if project.Spec.WorkerTemplate.DefaultConfigFiles == nil {
	// 	project.Spec.WorkerTemplate.DefaultConfigFiles = map[string]string{}
	// }

	return project
}
