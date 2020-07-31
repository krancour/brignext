package api

import (
	"encoding/json"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type SecretReference struct {
	Key string `json:"key,omitempty"`
}

func (s SecretReference) MarshalJSON() ([]byte, error) {
	type Alias SecretReference
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "SecretReference",
			},
			Alias: (Alias)(s),
		},
	)
}

type SecretReferenceList struct {
	Items []SecretReference `json:"items,omitempty"`
}

func (s SecretReferenceList) MarshalJSON() ([]byte, error) {
	type Alias SecretReferenceList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "SecretReferenceList",
			},
			Alias: (Alias)(s),
		},
	)
}
