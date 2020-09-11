package system

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/pkg/errors"
)

type RolesService interface {
	// TODO: This needs a function for listing available system roles
	// TODO: This needs a function for listing system role assignments
	Grant(
		ctx context.Context,
		roleAssignment authx.RoleAssignment,
	) error
	Revoke(
		ctx context.Context,
		roleAssignment authx.RoleAssignment,
	) error
}

type rolesService struct {
	authorize            authx.AuthorizeFn
	usersStore           authx.UsersStore
	serviceAccountsStore authx.ServiceAccountsStore
	rolesStore           authx.RolesStore
}

func NewRolesService(
	usersStore authx.UsersStore,
	serviceAccountsStore authx.ServiceAccountsStore,
	rolesStore authx.RolesStore,
) RolesService {
	return &rolesService{
		authorize:            authx.Authorize,
		usersStore:           usersStore,
		serviceAccountsStore: serviceAccountsStore,
		rolesStore:           rolesStore,
	}
}

func (s *rolesService) Grant(
	ctx context.Context,
	roleAssignment authx.RoleAssignment,
) error {
	if err := s.authorize(ctx, authx.RoleAdmin()); err != nil {
		return err
	}

	if roleAssignment.PrincipalType == authx.PrincipalTypeUser {
		// Make sure the User exists
		if _, err := s.usersStore.Get(ctx, roleAssignment.PrincipalID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving user %q from store",
				roleAssignment.PrincipalID,
			)
		}
	} else if roleAssignment.PrincipalType == authx.PrincipalTypeServiceAccount {
		// Make sure the ServiceAccount exists
		if _, err :=
			s.serviceAccountsStore.Get(ctx, roleAssignment.PrincipalID); err != nil {
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
	if err := s.rolesStore.Grant(
		ctx,
		roleAssignment.PrincipalType,
		roleAssignment.PrincipalID,
		authx.Role{
			Type:  authx.RoleTypeSystem,
			Name:  roleAssignment.Role,
			Scope: roleAssignment.Scope,
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error granting system role %q with scope %q to %s %q in store",
			roleAssignment.Role,
			roleAssignment.Scope,
			roleAssignment.PrincipalType,
			roleAssignment.PrincipalID,
		)
	}

	return nil
}

func (s *rolesService) Revoke(
	ctx context.Context,
	roleAssignment authx.RoleAssignment,
) error {
	if err := s.authorize(ctx, authx.RoleAdmin()); err != nil {
		return err
	}

	if roleAssignment.PrincipalType == authx.PrincipalTypeUser {
		// Make sure the User exists
		if _, err := s.usersStore.Get(ctx, roleAssignment.PrincipalID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving user %q from store",
				roleAssignment.PrincipalID,
			)
		}
	} else if roleAssignment.PrincipalType == authx.PrincipalTypeServiceAccount {
		// Make sure the ServiceAccount exists
		if _, err :=
			s.serviceAccountsStore.Get(ctx, roleAssignment.PrincipalID); err != nil {
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
	if err := s.rolesStore.Revoke(
		ctx,
		roleAssignment.PrincipalType,
		roleAssignment.PrincipalID,
		authx.Role{
			Type:  authx.RoleTypeSystem,
			Name:  roleAssignment.Role,
			Scope: roleAssignment.Scope,
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error revoking system role %q with scope %q for %s %q in store",
			roleAssignment.Role,
			roleAssignment.Scope,
			roleAssignment.PrincipalType,
			roleAssignment.PrincipalID,
		)
	}

	return nil
}
