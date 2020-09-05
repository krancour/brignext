package authx

import (
	"context"
)

type RolesStore interface {
	GrantRole(
		ctx context.Context,
		principalType PrincipalType,
		principalID string,
		roles ...Role,
	) error
	RevokeRole(
		ctx context.Context,
		principalType PrincipalType,
		principalID string,
		roles ...Role,
	) error
}
