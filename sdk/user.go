package sdk

import (
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type UserList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`
	Items         []User `json:"items"`
}

func NewUserList() UserList {
	return UserList{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "UserList",
		},
		Items: []User{},
	}
}

type User struct {
	meta.TypeMeta   `json:",inline" bson:",inline"`
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	Name            string     `json:"name" bson:"name"`
	Locked          *time.Time `json:"locked" bson:"locked"`
}

func NewUser(id, name string) User {
	return User{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "User",
		},
		ObjectMeta: meta.ObjectMeta{
			ID: id,
		},
		Name: name,
	}
}
