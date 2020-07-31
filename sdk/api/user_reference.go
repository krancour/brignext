package api

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type UserReference struct {
	meta.ObjectReferenceMeta `json:"metadata"`
	Name                     string     `json:"name"`
	Locked                   *time.Time `json:"locked"`
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
