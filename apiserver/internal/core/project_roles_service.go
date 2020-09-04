package core

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/pkg/errors"
)

type ProjectRolesService interface {
	// TODO: Implement this
	// ListUsers(context.Context) (authx.UserList, error)
	GrantToUser(
		ctx context.Context,
		projectID string,
		userID string,
		roleName string,
	) error
	RevokeFromUser(
		ctx context.Context,
		projectID string,
		userID string,
		roleName string,
	) error

	// TODO: Implement this
	// ListServiceAccounts(context.Context) (authx.UserList, error)
	GrantToServiceAccount(
		ctx context.Context,
		projectID string,
		serviceAccountID string,
		roleName string,
	) error
	RevokeFromServiceAccount(
		ctx context.Context,
		projectID string,
		serviceAccountID string,
		roleName string,
	) error
}

type projectRolesService struct {
	authorize            authx.AuthorizeFn
	projectsStore        ProjectsStore
	usersStore           authx.UsersStore
	serviceAccountsStore authx.ServiceAccountsStore
	rolesStore           authx.RolesStore
}

func NewProjectRolesService(
	projectsStore ProjectsStore,
	usersStore authx.UsersStore,
	serviceAccountsStore authx.ServiceAccountsStore,
	rolesStore authx.RolesStore,
) ProjectRolesService {
	return &projectRolesService{
		authorize:            authx.Authorize,
		projectsStore:        projectsStore,
		usersStore:           usersStore,
		serviceAccountsStore: serviceAccountsStore,
		rolesStore:           rolesStore,
	}
}

func (p *projectRolesService) GrantToUser(
	ctx context.Context,
	projectID string,
	userID string,
	roleName string,
) error {
	if err := p.authorize(ctx, authx.RoleProjectAdmin(projectID)); err != nil {
		return err
	}

	// Make sure the project exists
	_, err := p.projectsStore.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}

	// Make sure the User exists
	if _, err := p.usersStore.Get(ctx, userID); err != nil {
		return errors.Wrapf(err, "error retrieving user %q from store", userID)
	}

	// Give them the Role
	return p.rolesStore.GrantToUser(
		ctx,
		userID,
		authx.Role{
			Type:  "PROJECT",
			Name:  roleName,
			Scope: projectID,
		})
}

func (p *projectRolesService) RevokeFromUser(
	ctx context.Context,
	projectID string,
	userID string,
	roleName string,
) error {
	if err := p.authorize(ctx, authx.RoleProjectAdmin(projectID)); err != nil {
		return err
	}

	// Make sure the project exists
	_, err := p.projectsStore.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}

	// Make sure the User exists
	if _, err := p.usersStore.Get(ctx, userID); err != nil {
		return errors.Wrapf(err, "error retrieving user %q from store", userID)
	}

	// Revoke the Role
	return p.rolesStore.RevokeFromUser(
		ctx,
		userID,
		authx.Role{
			Type:  "PROJECT",
			Name:  roleName,
			Scope: projectID,
		},
	)
}

func (p *projectRolesService) GrantToServiceAccount(
	ctx context.Context,
	projectID string,
	serviceAccountID string,
	roleName string,
) error {
	if err := p.authorize(ctx, authx.RoleProjectAdmin(projectID)); err != nil {
		return err
	}

	// Make sure the project exists
	_, err := p.projectsStore.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}

	// Make sure the ServiceAccount exists
	if _, err := p.serviceAccountsStore.Get(ctx, serviceAccountID); err != nil {
		return errors.Wrapf(
			err,
			"error retrieving service account %q from store",
			serviceAccountID,
		)
	}

	// Give it the Role
	return p.rolesStore.GrantToServiceAccount(
		ctx,
		serviceAccountID,
		authx.Role{
			Type:  "PROJECT",
			Name:  roleName,
			Scope: projectID,
		})
}

func (p *projectRolesService) RevokeFromServiceAccount(
	ctx context.Context,
	projectID string,
	serviceAccountID string,
	roleName string,
) error {
	if err := p.authorize(ctx, authx.RoleProjectAdmin(projectID)); err != nil {
		return err
	}

	// Make sure the project exists
	_, err := p.projectsStore.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}

	// Make sure the ServiceAccount exists
	if _, err := p.serviceAccountsStore.Get(ctx, serviceAccountID); err != nil {
		return errors.Wrapf(
			err,
			"error retrieving service account %q from store",
			serviceAccountID,
		)
	}

	// Revoke the Role
	return p.rolesStore.RevokeFromServiceAccount(
		ctx,
		serviceAccountID,
		authx.Role{
			Type:  "PROJECT",
			Name:  roleName,
			Scope: projectID,
		},
	)
}
