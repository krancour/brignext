package sdk

import (
	"encoding/json"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type UserSessionAuthDetails struct {
	OAuth2State string `json:"oauth2State"`
	AuthURL     string `json:"authURL"`
	Token       string `json:"token"`
}

func (u UserSessionAuthDetails) MarshalJSON() ([]byte, error) {
	type Alias UserSessionAuthDetails
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "UserSessionAuthDetails",
			},
			Alias: (Alias)(u),
		},
	)
}
