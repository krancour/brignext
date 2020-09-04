package authx

import (
	"context"
)

type RolesStore interface {
	GrantToUser(ctx context.Context, userID string, roles ...Role) error
	RevokeFromUser(ctx context.Context, userID string, roles ...Role) error

	GrantToServiceAccount(
		ctx context.Context,
		serviceAccountID string,
		roles ...Role,
	) error
	RevokeFromServiceAccount(
		ctx context.Context,
		serviceAccountID string,
		roles ...Role,
	) error
}
