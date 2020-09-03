package authx

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

// User represents a (human) BrigNext user.
type User struct {
	// ObjectMeta encapsulates User metadata.
	meta.ObjectMeta `json:"metadata"`
	// Name is the given name and surname of the User.
	Name string `json:"name,omitempty"`
	// Locked indicates when the User has been locked out of the system by
	// an administrator. If this field's value is nil, the User can be presumed
	// NOT to be locked.
	Locked *time.Time `json:"locked,omitempty"`
	Roles  []Role     `json:"roles,omitempty"`
}

// MarshalJSON amends User instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
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
