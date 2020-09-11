package authx

import (
	"encoding/json"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
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