package core

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/pkg/errors"
)

type ProjectRolesService interface {
	// TODO: This needs a function for listing available project roles
	// TODO: This needs a function for listing role assignments by project
	Grant(
		ctx context.Context,
		projectID string,
		roleAssignment authx.RoleAssignment,
	) error
	Revoke(
		ctx context.Context,
		projectID string,
		roleAssignment authx.RoleAssignment,
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

func (p *projectRolesService) Grant(
	ctx context.Context,
	projectID string,
	roleAssignment authx.RoleAssignment,
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

	if roleAssignment.PrincipalType == authx.PrincipalTypeUser {
		// Make sure the User exists
		if _, err := p.usersStore.Get(ctx, roleAssignment.PrincipalID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving user %q from store",
				roleAssignment.PrincipalID,
			)
		}
	} else if roleAssignment.PrincipalType == authx.PrincipalTypeServiceAccount {
		// Make sure the ServiceAccount exists
		if _, err :=
			p.serviceAccountsStore.Get(ctx, roleAssignment.PrincipalID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving service account %q from store",
				roleAssignment.PrincipalID,
			)
		}
	} else {
		return nil
	}

	// Give them the Role
	if err := p.rolesStore.Grant(
		ctx,
		roleAssignment.PrincipalType,
		roleAssignment.PrincipalID,
		authx.Role{
			Type:  authx.RoleTypeProject,
			Name:  roleAssignment.Role,
			Scope: projectID,
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error granting project %q role %q to %s %q in store",
			projectID,
			roleAssignment.Role,
			roleAssignment.PrincipalType,
			roleAssignment.PrincipalID,
		)
	}

	return nil
}

func (p *projectRolesService) Revoke(
	ctx context.Context,
	projectID string,
	roleAssignment authx.RoleAssignment,
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

	if roleAssignment.PrincipalType == authx.PrincipalTypeUser {
		// Make sure the User exists
		if _, err := p.usersStore.Get(ctx, roleAssignment.PrincipalID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving user %q from store",
				roleAssignment.PrincipalID,
			)
		}
	} else if roleAssignment.PrincipalType == authx.PrincipalTypeServiceAccount {
		// Make sure the ServiceAccount exists
		if _, err :=
			p.serviceAccountsStore.Get(ctx, roleAssignment.PrincipalID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving service account %q from store",
				roleAssignment.PrincipalID,
			)
		}
	} else {
		return nil
	}

	// Revoke the Role
	if err := p.rolesStore.Revoke(
		ctx,
		roleAssignment.PrincipalType,
		roleAssignment.PrincipalID,
		authx.Role{
			Type:  authx.RoleTypeProject,
			Name:  roleAssignment.Role,
			Scope: projectID,
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error revoking project %q role %q for %s %q in store",
			projectID,
			roleAssignment.Role,
			roleAssignment.PrincipalType,
			roleAssignment.PrincipalID,
		)
	}
	return nil
}
