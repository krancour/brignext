package sdk

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
)

type User struct {
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	Name            string     `json:"name" bson:"name"`
	Locked          *time.Time `json:"locked" bson:"locked"`
}

func (u User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "User",
			},
			Alias: (Alias)(u),
		},
	)
}

type UserReference struct {
	meta.ObjectReferenceMeta `json:"metadata" bson:",inline"`
	Name                     string     `json:"name" bson:"name"`
	Locked                   *time.Time `json:"locked" bson:"locked"`
}

func (u UserReference) MarshalJSON() ([]byte, error) {
	type Alias UserReference
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "UserReference",
			},
			Alias: (Alias)(u),
		},
	)
}

type UserReferenceList struct {
	Items []UserReference `json:"items"`
}

func (u UserReferenceList) MarshalJSON() ([]byte, error) {
	type Alias UserReferenceList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "UserReferenceList",
			},
			Alias: (Alias)(u),
		},
	)
}
