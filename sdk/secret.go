package sdk

import (
	"encoding/json"

	"github.com/krancour/brignext/v2/sdk/meta"
)

// Secret represents Project-level sensitive information.
type Secret struct {
	// Key is a key by which the secret can referred.
	Key string `json:"key"`
	// Value is the sensitive information.
	Value string `json:"value"`
}

// MarshalJSON amends Secret instances with type metadata so that clients do not
// need to be concerned with the tedium of doing so.
func (s Secret) MarshalJSON() ([]byte, error) {
	type Alias Secret
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "Secret",
			},
			Alias: (Alias)(s),
		},
	)
}
