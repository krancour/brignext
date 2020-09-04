package system

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/pkg/errors"
)

type SystemRolesService interface {
	// TODO: Implement this
	// ListUsers(context.Context) (authx.UserList, error)
	GrantToUser(
		ctx context.Context,
		userID string,
		roleName string,
	) error
	RevokeFromUser(
		ctx context.Context,
		userID string,
		roleName string,
	) error

	// TODO: Implement this
	// ListServiceAccounts(context.Context) (authx.UserList, error)
	GrantToServiceAccount(
		ctx context.Context,
		serviceAccountID string,
		roleName string,
	) error
	RevokeFromServiceAccount(
		ctx context.Context,
		serviceAccountID string,
		roleName string,
	) error
}

type systemRolesService struct {
	authorize            authx.AuthorizeFn
	usersStore           authx.UsersStore
	serviceAccountsStore authx.ServiceAccountsStore
}

func NewSystemRolesService(
	usersStore authx.UsersStore,
	serviceAccountsStore authx.ServiceAccountsStore,
) SystemRolesService {
	return &systemRolesService{
		authorize:            authx.Authorize,
		usersStore:           usersStore,
		serviceAccountsStore: serviceAccountsStore,
	}
}

func (s *systemRolesService) GrantToUser(
	ctx context.Context,
	userID string,
	roleName string,
) error {
	if err := s.authorize(ctx, authx.RoleAdmin()); err != nil {
		return err
	}

	// Make sure the User exists
	if _, err := s.usersStore.Get(ctx, userID); err != nil {
		return errors.Wrapf(err, "error retrieving user %q from store", userID)
	}

	// Give them the Role
	return s.usersStore.GrantRole(
		ctx,
		userID,
		authx.Role{
			Type:  "SYSTEM",
			Name:  roleName,
			Scope: authx.RoleScopeGlobal,
		},
	)
}

func (s *systemRolesService) RevokeFromUser(
	ctx context.Context,
	userID string,
	roleName string,
) error {
	if err := s.authorize(ctx, authx.RoleAdmin()); err != nil {
		return err
	}

	// Make sure the User exists
	if _, err := s.usersStore.Get(ctx, userID); err != nil {
		return errors.Wrapf(err, "error retrieving user %q from store", userID)
	}

	// Revoke the Role
	return s.usersStore.RevokeRole(
		ctx,
		userID,
		authx.Role{
			Type:  "SYSTEM",
			Name:  roleName,
			Scope: authx.RoleScopeGlobal,
		},
	)
}

func (s *systemRolesService) GrantToServiceAccount(
	ctx context.Context,
	serviceAccountID string,
	roleName string,
) error {
	if err := s.authorize(ctx, authx.RoleAdmin()); err != nil {
		return err
	}

	// Make sure the ServiceAccount exists
	if _, err := s.serviceAccountsStore.Get(ctx, serviceAccountID); err != nil {
		return errors.Wrapf(
			err,
			"error retrieving service account %q from store",
			serviceAccountID,
		)
	}

	// Give it the Role
	return s.serviceAccountsStore.GrantRole(
		ctx,
		serviceAccountID,
		authx.Role{
			Type:  "SYSTEM",
			Name:  roleName,
			Scope: authx.RoleScopeGlobal,
		},
	)
}

func (s *systemRolesService) RevokeFromServiceAccount(
	ctx context.Context,
	serviceAccountID string,
	roleName string,
) error {
	if err := s.authorize(ctx, authx.RoleAdmin()); err != nil {
		return err
	}

	// Make sure the ServiceAccount exists
	if _, err := s.serviceAccountsStore.Get(ctx, serviceAccountID); err != nil {
		return errors.Wrapf(
			err,
			"error retrieving service account %q from store",
			serviceAccountID,
		)
	}

	// Revoke the Role
	return s.serviceAccountsStore.RevokeRole(
		ctx,
		serviceAccountID,
		authx.Role{
			Type:  "SYSTEM",
			Name:  roleName,
			Scope: authx.RoleScopeGlobal,
		},
	)
}
