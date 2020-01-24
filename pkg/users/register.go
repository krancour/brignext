package users

import (
	context "context"

	"github.com/pkg/errors"
)

func (u *usersServer) Register(
	_ context.Context,
	req *RegistrationRequest,
) (*RegistrationResponse, error) {
	// TODO: We should do some kind of validation!

	resp := &RegistrationResponse{}

	if err := u.userStore.CreateUser(req.Username, req.Password); err != nil {
		return resp, errors.Wrapf(err, "error creating user %q", req.Username)
	}

	return resp, nil
}
