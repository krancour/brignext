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

// SecretListOptions represents useful filter criteria when selecting multiple
// Secrets for API group operations like list.
type SecretListOptions struct {
	Continue string // TODO: Clean this up
	Limit    int64  // TODO: Clean this up
}

// SecretList is an ordered and pageable list of Secrets.
type SecretList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Secrets.
	Items []Secret `json:"items,omitempty"`
}

func (s SecretList) Len() int {
	return len(s.Items)
}

func (s SecretList) Swap(i, j int) {
	s.Items[i], s.Items[j] = s.Items[j], s.Items[i]
}

func (s SecretList) Less(i, j int) bool {
	return s.Items[i].Key < s.Items[j].Key
}

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
