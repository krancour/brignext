package users

import "github.com/krancour/brignext/pkg/storage"

type usersServer struct {
	userStore storage.UserStore
}

func NewServer(
	userStore storage.UserStore,
) UsersServer {
	return &usersServer{
		userStore: userStore,
	}
}
