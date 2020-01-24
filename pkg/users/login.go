package users

import (
	context "context"

	"github.com/krancour/brignext/pkg/auth"
	"github.com/pkg/errors"
)

func (u *usersServer) Login(
	ctx context.Context,
	_ *LoginRequest,
) (*LoginResponse, error) {

	resp := &LoginResponse{}

	user := auth.UserFromContext(ctx)
	if user == nil {
		return resp, errors.New(
			"auth malfunction-- no user found in request context",
		)
	}

	token, err := u.userStore.CreateUserToken(user.Username)
	if err != nil {
		return resp, errors.Wrapf(err, "error logging in user %q", user.Username)
	}

	resp.Token = token

	return resp, nil
}
