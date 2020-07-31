package sdk

import (
	"encoding/json"

	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
)

type Secret struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

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

type SecretReference struct {
	Key string `json:"key"`
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
	Items []SecretReference `json:"items"`
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
