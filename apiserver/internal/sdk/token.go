package sdk

import (
	"encoding/json"

	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
)

type Token struct {
	Value string `json:"value" bson:"value"`
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
