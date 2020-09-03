package api

import (
	"encoding/json"

	"github.com/krancour/brignext/v2/sdk/core"
	"github.com/krancour/brignext/v2/sdk/meta"
)

// SecretList is an ordered and pageable list of Secrets.
type SecretList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Secrets.
	Items []core.Secret `json:"items,omitempty"`
}

// MarshalJSON amends SecretList instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (s SecretList) MarshalJSON() ([]byte, error) {
	type Alias SecretList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "SecretList",
			},
			Alias: (Alias)(s),
		},
	)
}
