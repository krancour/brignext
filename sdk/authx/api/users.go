package api

import (
	"encoding/json"

	"github.com/brigadecore/brigade/v2/sdk/authx"
	"github.com/brigadecore/brigade/v2/sdk/meta"
)

// UsersSelector represents useful filter criteria when selecting multiple Users
// for API group operations like list. It currently has no fields, but exists
// for future expansion.
type UsersSelector struct{}

// UserList is an ordered and pageable list of Users.
type UserList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Users.
	Items []authx.User `json:"items,omitempty"`
}

// MarshalJSON amends UserList instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (u UserList) MarshalJSON() ([]byte, error) {
	type Alias UserList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "UserList",
			},
			Alias: (Alias)(u),
		},
	)
}
