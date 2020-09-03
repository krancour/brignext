package system

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/pkg/errors"
)

type Service interface {
	// TODO: Implement this
	// ListUsers(context.Context) (authx.UserList, error)
	GrantRoleToUser(
		ctx context.Context,
		userID string,
		roleName string,
	) error
	RevokeRoleFromUser(
		ctx context.Context,
		userID string,
		roleName string,
	) error

	// TODO: Implement this
	// ListServiceAccounts(context.Context) (authx.UserList, error)
	GrantRoleToServiceAccount(
		ctx context.Context,
		serviceAccountID string,
		roleName string,
	) error
	RevokeRoleFromServiceAccount(
		ctx context.Context,
		serviceAccountID string,
		roleName string,
	) error
}

type service struct {
	authorize            authx.AuthorizeFn
	usersStore           authx.UsersStore
	serviceAccountsStore authx.ServiceAccountsStore
}

func NewService(
	usersStore authx.UsersStore,
	serviceAccountsStore authx.ServiceAccountsStore,
) Service {
	return &service{
		authorize:            authx.Authorize,
		usersStore:           usersStore,
		serviceAccountsStore: serviceAccountsStore,
	}
}

func (s *service) GrantRoleToUser(
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

func (s *service) RevokeRoleFromUser(
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

func (s *service) GrantRoleToServiceAccount(
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

func (s *service) RevokeRoleFromServiceAccount(
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
