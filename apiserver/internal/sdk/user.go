package sdk

import (
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
	"go.mongodb.org/mongo-driver/bson"
)

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

type UserReference struct {
	meta.TypeMeta            `json:",inline"`
	meta.ObjectReferenceMeta `json:"metadata" bson:",inline"`
	Name                     string     `json:"name" bson:"name"`
	Locked                   *time.Time `json:"locked" bson:"locked"`
}

func (u *UserReference) UnmarshalBSON(bytes []byte) error {
	type UserReferenceAlias UserReference
	if err := bson.Unmarshal(
		bytes,
		&struct {
			*UserReferenceAlias `bson:",inline"`
		}{
			UserReferenceAlias: (*UserReferenceAlias)(u),
		},
	); err != nil {
		return err
	}
	u.TypeMeta = meta.TypeMeta{
		APIVersion: meta.APIVersion,
		Kind:       "UserReference",
	}
	return nil
}

type UserReferenceList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`
	Items         []UserReference `json:"items"`
}

func NewUserReferenceList() UserReferenceList {
	return UserReferenceList{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "UserReferenceList",
		},
		ListMeta: meta.ListMeta{},
		Items:    []UserReference{},
	}
}
