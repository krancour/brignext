package sdk

import (
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type User struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata"`
	Name            string     `json:"name"`
	Locked          *time.Time `json:"locked"`
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

type UserReference struct {
	meta.TypeMeta            `json:",inline"`
	meta.ObjectReferenceMeta `json:"metadata"`
	Name                     string     `json:"name"`
	Locked                   *time.Time `json:"locked"`
}

type UserReferenceList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`
	Items         []UserReference `json:"items"`
}
