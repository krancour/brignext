package projects

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/authx/serviceaccounts"
	"github.com/krancour/brignext/v2/apiserver/internal/authx/users"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// Service is the specialized interface for managing Projects. It's decoupled
// from underlying technology choices (e.g. data store, message bus, etc.) to
// keep business logic reusable and consistent while the underlying tech stack
// remains free to change.
type Service interface {
	// Create creates a new Project.
	Create(context.Context, core.Project) (core.Project, error)
	// List returns a ProjectList, with its Items (Projects) ordered
	// alphabetically by Project ID.
	List(
		context.Context,
		core.ProjectsSelector,
		meta.ListOptions,
	) (core.ProjectList, error)
	// Get retrieves a single Project specified by its identifier.
	Get(context.Context, string) (core.Project, error)
	// Update updates an existing Project.
	Update(context.Context, core.Project) (core.Project, error)
	// Delete deletes a single Project specified by its identifier.
	Delete(context.Context, string) error

	// ListSecrets returns a SecretList who Items (Secrets) contain Keys only and
	// not Values (all Value fields are empty). i.e. Once a secret is set, end
	// clients are unable to retrieve values.
	ListSecrets(
		ctx context.Context,
		projectID string,
		opts meta.ListOptions,
	) (core.SecretList, error)
	// SetSecret set the value of a new Secret or updates the value of an existing
	// Secret.
	SetSecret(
		ctx context.Context,
		projectID string,
		secret core.Secret,
	) error
	// UnsetSecret clears the value of an existing Secret.
	UnsetSecret(ctx context.Context, projectID string, key string) error

	GrantRole(
		ctx context.Context,
		projectID string,
		roleAssignment authx.RoleAssignment,
	) error
	RevokeRole(
		ctx context.Context,
		projectID string,
		roleAssignment authx.RoleAssignment,
	) error
}

type service struct {
	authorize            authx.AuthorizeFn
	store                Store
	usersStore           users.Store
	serviceAccountsStore serviceaccounts.Store
	scheduler            Scheduler
}

// NewService returns a specialized interface for managing Projects.
func NewService(
	store Store,
	usersStore users.Store,
	serviceAccountsStore serviceaccounts.Store,
	scheduler Scheduler,
) Service {
	return &service{
		authorize:            authx.Authorize,
		store:                store,
		usersStore:           usersStore,
		serviceAccountsStore: serviceAccountsStore,
		scheduler:            scheduler,
	}
}

func (s *service) Create(
	ctx context.Context,
	project core.Project,
) (core.Project, error) {
	if err := s.authorize(ctx, authx.RoleProjectCreator()); err != nil {
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
	if project, err = s.scheduler.PreCreate(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error pre-creating project %q in the scheduler",
			project.ID,
		)
	}

	if err = s.store.Create(ctx, project); err != nil {
		return project,
			errors.Wrapf(err, "error storing new project %q", project.ID)
	}
	if err = s.scheduler.Create(ctx, project); err != nil {
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
		if err := s.usersStore.GrantRole(
			ctx,
			user.ID,
			roles...,
		); err != nil {
			return core.Project{}, errors.Wrapf(
				err,
				"error storing project %q roles for user %q",
				project.ID,
				user.ID,
			)
		}
	} else if serviceAccount, ok := principal.(*authx.ServiceAccount); ok {
		if err := s.serviceAccountsStore.GrantRole(
			ctx,
			serviceAccount.ID,
			roles...,
		); err != nil {
			return core.Project{}, errors.Wrapf(
				err,
				"error storing project %q roles for service account %q",
				project.ID,
				serviceAccount.ID,
			)
		}
	}

	return project, nil
}

func (s *service) List(
	ctx context.Context,
	selector core.ProjectsSelector,
	opts meta.ListOptions,
) (core.ProjectList, error) {
	if err := s.authorize(ctx, authx.RoleReader()); err != nil {
		return core.ProjectList{}, err
	}

	if opts.Limit == 0 {
		opts.Limit = 20
	}
	projects, err := s.store.List(ctx, selector, opts)
	if err != nil {
		return projects, errors.Wrap(err, "error retrieving projects from store")
	}
	return projects, nil
}

func (s *service) Get(
	ctx context.Context,
	id string,
) (core.Project, error) {
	if err := s.authorize(ctx, authx.RoleReader()); err != nil {
		return core.Project{}, err
	}

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
	project core.Project,
) (core.Project, error) {
	if err := s.authorize(
		ctx,
		authx.RoleProjectDeveloper(project.ID),
	); err != nil {
		return core.Project{}, err
	}

	// Let the scheduler update scheduler-specific details before we persist.
	var err error
	if project, err = s.scheduler.PreUpdate(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error pre-updating project %q in the scheduler",
			project.ID,
		)
	}

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
	if err := s.authorize(ctx, authx.RoleProjectAdmin(id)); err != nil {
		return err
	}

	project, err := s.store.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving project %q from store", id)
	}

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
	opts meta.ListOptions,
) (core.SecretList, error) {
	if err := s.authorize(ctx, authx.RoleReader()); err != nil {
		return core.SecretList{}, err
	}

	secrets := core.SecretList{}
	project, err := s.store.Get(ctx, projectID)
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

func (s *service) SetSecret(
	ctx context.Context,
	projectID string,
	secret core.Secret,
) error {
	if err := s.authorize(ctx, authx.RoleProjectAdmin(projectID)); err != nil {
		return err
	}

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
	if err := s.authorize(ctx, authx.RoleProjectAdmin(projectID)); err != nil {
		return err
	}

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

// TODO: Implement this
func (s *service) GrantRole(
	ctx context.Context,
	projectID string,
	roleAssignment authx.RoleAssignment,
) error {
	if err := s.authorize(ctx, authx.RoleProjectAdmin(projectID)); err != nil {
		return err
	}

	if (roleAssignment.UserID == "" && roleAssignment.ServiceAccountID == "") ||
		(roleAssignment.UserID != "" && roleAssignment.ServiceAccountID != "") {
		return &core.ErrBadRequest{} // TODO: Add more context
	}

	// Make sure the project exists
	_, err := s.store.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}

	role := authx.Role{
		Type:  "PROJECT",
		Name:  roleAssignment.Role,
		Scope: projectID,
	}

	if roleAssignment.UserID != "" {
		// Make sure the user exists
		if _, err := s.usersStore.Get(ctx, roleAssignment.UserID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving user %q from store",
				roleAssignment.UserID,
			)
		}
		// Give them the role
		return s.usersStore.GrantRole(ctx, roleAssignment.UserID, role)
	}

	// Make sure the ServiceAccount exists
	if _, err := s.serviceAccountsStore.Get(
		ctx,
		roleAssignment.ServiceAccountID,
	); err != nil {
		return errors.Wrapf(
			err,
			"error retrieving service account %q from store",
			roleAssignment.ServiceAccountID,
		)
	}
	// Give it the role
	return s.serviceAccountsStore.GrantRole(ctx, roleAssignment.UserID, role)
}

// TODO: Implement this
func (s *service) RevokeRole(
	ctx context.Context,
	projectID string,
	roleAssignment authx.RoleAssignment,
) error {
	if err := s.authorize(ctx, authx.RoleProjectAdmin(projectID)); err != nil {
		return err
	}

	if (roleAssignment.UserID == "" && roleAssignment.ServiceAccountID == "") ||
		(roleAssignment.UserID != "" && roleAssignment.ServiceAccountID != "") {
		return &core.ErrBadRequest{} // TODO: Add more context
	}

	// Make sure the project exists
	_, err := s.store.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}

	role := authx.Role{
		Type:  "PROJECT",
		Name:  roleAssignment.Role,
		Scope: projectID,
	}

	if roleAssignment.UserID != "" {
		// Make sure the user exists
		if _, err := s.usersStore.Get(ctx, roleAssignment.UserID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving user %q from store",
				roleAssignment.UserID,
			)
		}
		// Revoke the role
		return s.usersStore.RevokeRole(ctx, roleAssignment.UserID, role)
	}

	// Make sure the ServiceAccount exists
	if _, err := s.serviceAccountsStore.Get(
		ctx,
		roleAssignment.ServiceAccountID,
	); err != nil {
		return errors.Wrapf(
			err,
			"error retrieving service account %q from store",
			roleAssignment.ServiceAccountID,
		)
	}
	// Revoke the role
	return s.serviceAccountsStore.RevokeRole(ctx, roleAssignment.UserID, role)
}
