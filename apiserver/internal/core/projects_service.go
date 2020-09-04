package core

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// ProjectsService is the specialized interface for managing Projects. It's decoupled
// from underlying technology choices (e.g. data store, message bus, etc.) to
// keep business logic reusable and consistent while the underlying tech stack
// remains free to change.
type ProjectsService interface {
	// Create creates a new Project.
	Create(context.Context, Project) (Project, error)
	// List returns a ProjectList, with its Items (Projects) ordered
	// alphabetically by Project ID.
	List(
		context.Context,
		ProjectsSelector,
		meta.ListOptions,
	) (ProjectList, error)
	// Get retrieves a single Project specified by its identifier.
	Get(context.Context, string) (Project, error)
	// Update updates an existing Project.
	Update(context.Context, Project) (Project, error)
	// Delete deletes a single Project specified by its identifier.
	Delete(context.Context, string) error
}

type projectsService struct {
	authorize            authx.AuthorizeFn
	store                ProjectsStore
	usersStore           authx.UsersStore
	serviceAccountsStore authx.ServiceAccountsStore
	scheduler            ProjectsScheduler
}

// NewProjectsService returns a specialized interface for managing Projects.
func NewProjectsService(
	store ProjectsStore,
	usersStore authx.UsersStore,
	serviceAccountsStore authx.ServiceAccountsStore,
	scheduler ProjectsScheduler,
) ProjectsService {
	return &projectsService{
		authorize:            authx.Authorize,
		store:                store,
		usersStore:           usersStore,
		serviceAccountsStore: serviceAccountsStore,
		scheduler:            scheduler,
	}
}

func (p *projectsService) Create(
	ctx context.Context,
	project Project,
) (Project, error) {
	if err := p.authorize(ctx, authx.RoleProjectCreator()); err != nil {
		return project, err
	}

	now := time.Now()
	project.Created = &now

	// TODO: The principal that created this should automatically be a project
	// admin, developer, and user... but how can we do that without creating a
	// cyclic dependency? (UserService and ServiceAccountService already depend
	// on this package because they use the Project store to check the validity
	// of a project scope.)

	// Let the scheduler add scheduler-specific details before we persist.
	var err error
	if project, err = p.scheduler.PreCreate(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error pre-creating project %q in the scheduler",
			project.ID,
		)
	}

	if err = p.store.Create(ctx, project); err != nil {
		return project,
			errors.Wrapf(err, "error storing new project %q", project.ID)
	}
	if err = p.scheduler.Create(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating project %q in the scheduler",
			project.ID,
		)
	}

	// Assign roles to the principal who created the project...
	principal := authx.PincipalFromContext(ctx)
	roles := []authx.Role{
		authx.RoleProjectAdmin(project.ID),
		authx.RoleProjectDeveloper(project.ID),
		authx.RoleProjectUser(project.ID),
	}
	if user, ok := principal.(*authx.User); ok {
		if err := p.usersStore.GrantRole(
			ctx,
			user.ID,
			roles...,
		); err != nil {
			return Project{}, errors.Wrapf(
				err,
				"error storing project %q roles for user %q",
				project.ID,
				user.ID,
			)
		}
	} else if serviceAccount, ok := principal.(*authx.ServiceAccount); ok {
		if err := p.serviceAccountsStore.GrantRole(
			ctx,
			serviceAccount.ID,
			roles...,
		); err != nil {
			return Project{}, errors.Wrapf(
				err,
				"error storing project %q roles for service account %q",
				project.ID,
				serviceAccount.ID,
			)
		}
	}

	return project, nil
}

func (p *projectsService) List(
	ctx context.Context,
	selector ProjectsSelector,
	opts meta.ListOptions,
) (ProjectList, error) {
	if err := p.authorize(ctx, authx.RoleReader()); err != nil {
		return ProjectList{}, err
	}

	if opts.Limit == 0 {
		opts.Limit = 20
	}
	projects, err := p.store.List(ctx, selector, opts)
	if err != nil {
		return projects, errors.Wrap(err, "error retrieving projects from store")
	}
	return projects, nil
}

func (p *projectsService) Get(
	ctx context.Context,
	id string,
) (Project, error) {
	if err := p.authorize(ctx, authx.RoleReader()); err != nil {
		return Project{}, err
	}

	project, err := p.store.Get(ctx, id)
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
	project Project,
) (Project, error) {
	if err := p.authorize(
		ctx,
		authx.RoleProjectDeveloper(project.ID),
	); err != nil {
		return Project{}, err
	}

	// Let the scheduler update scheduler-specific details before we persist.
	var err error
	if project, err = p.scheduler.PreUpdate(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error pre-updating project %q in the scheduler",
			project.ID,
		)
	}

	if err = p.store.Update(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error updating project %q in store",
			project.ID,
		)
	}
	if err = p.scheduler.Update(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error updating project %q in the scheduler",
			project.ID,
		)
	}
	return project, nil
}

func (p *projectsService) Delete(ctx context.Context, id string) error {
	if err := p.authorize(ctx, authx.RoleProjectAdmin(id)); err != nil {
		return err
	}

	project, err := p.store.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving project %q from store", id)
	}

	if err := p.store.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error removing project %q from store", id)
	}
	if err := p.scheduler.Delete(ctx, project); err != nil {
		return errors.Wrapf(
			err,
			"error deleting project %q from scheduler",
			id,
		)
	}
	return nil
}
