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
	projectsStore        ProjectsStore
	usersStore           authx.UsersStore
	serviceAccountsStore authx.ServiceAccountsStore
	rolesStore           authx.RolesStore
	substrate            Substrate
}

// NewProjectsService returns a specialized interface for managing Projects.
func NewProjectsService(
	projectsStore ProjectsStore,
	usersStore authx.UsersStore,
	serviceAccountsStore authx.ServiceAccountsStore,
	rolesStore authx.RolesStore,
	substrate Substrate,
) ProjectsService {
	return &projectsService{
		authorize:            authx.Authorize,
		projectsStore:        projectsStore,
		usersStore:           usersStore,
		serviceAccountsStore: serviceAccountsStore,
		rolesStore:           rolesStore,
		substrate:            substrate,
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

	// Add substrate-specific details before we persist.
	var err error
	if project, err = p.substrate.PreCreateProject(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error pre-creating project %q on the substrate",
			project.ID,
		)
	}

	if err = p.projectsStore.Create(ctx, project); err != nil {
		return project,
			errors.Wrapf(err, "error storing new project %q", project.ID)
	}
	if err = p.substrate.CreateProject(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating project %q on the substrate",
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
		if err := p.rolesStore.GrantToUser(
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
		if err := p.rolesStore.GrantToServiceAccount(
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
	projects, err := p.projectsStore.List(ctx, selector, opts)
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

	project, err := p.projectsStore.Get(ctx, id)
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
	updatedProject Project,
) (Project, error) {
	if err := p.authorize(
		ctx,
		authx.RoleProjectDeveloper(updatedProject.ID),
	); err != nil {
		return Project{}, err
	}

	var err error
	oldProject, err := p.projectsStore.Get(ctx, updatedProject.ID)
	if err != nil {
		return updatedProject, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			updatedProject.ID,
		)
	}

	// Update substrate-specific details before we persist.
	if updatedProject, err = p.substrate.PreUpdateProject(ctx, oldProject, updatedProject); err != nil {
		return updatedProject, errors.Wrapf(
			err,
			"error pre-updating project %q on the substrate",
			updatedProject.ID,
		)
	}

	if err = p.projectsStore.Update(ctx, updatedProject); err != nil {
		return updatedProject, errors.Wrapf(
			err,
			"error updating project %q in store",
			updatedProject.ID,
		)
	}
	if err = p.substrate.UpdateProject(ctx, oldProject, updatedProject); err != nil {
		return updatedProject, errors.Wrapf(
			err,
			"error updating project %q on the substrate",
			updatedProject.ID,
		)
	}
	return updatedProject, nil
}

func (p *projectsService) Delete(ctx context.Context, id string) error {
	if err := p.authorize(ctx, authx.RoleProjectAdmin(id)); err != nil {
		return err
	}

	project, err := p.projectsStore.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving project %q from store", id)
	}

	if err := p.projectsStore.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error removing project %q from store", id)
	}
	if err := p.substrate.DeleteProject(ctx, project); err != nil {
		return errors.Wrapf(
			err,
			"error deleting project %q from substrate",
			id,
		)
	}
	return nil
}
