package api

import (
	"encoding/json"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type Token struct {
	Value string `json:"value"`
}

func (t Token) MarshalJSON() ([]byte, error) {
	type Alias Token
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "Token",
			},
			Alias: (Alias)(t),
		},
	)
}
