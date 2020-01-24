package users

import (
	context "context"

	"github.com/krancour/brignext/pkg/auth"
	"github.com/pkg/errors"
)

func (u *usersServer) Logout(
	ctx context.Context,
	_ *LogoutRequest,
) (*LogoutResponse, error) {

	resp := &LogoutResponse{}

	token := auth.UserTokenFromContext(ctx)
	if token == "" {
		return resp, errors.New(
			"auth malfunction-- no user token found in request context",
		)
	}

	if err := u.userStore.DeleteUserToken(token); err != nil {
		return resp, errors.Wrap(err, "error deleting token [REDACTED]")
	}

	return resp, nil
}
