package brignext

import "github.com/krancour/brignext/v2/meta"

type UserSessionAuthDetails struct {
	meta.TypeMeta `json:",inline" bson:",inline"`
	OAuth2State   string `json:"oauth2State"`
	AuthURL       string `json:"authURL"`
	Token         string `json:"token"`
}

func NewUserSessionAuthDetails(
	oauth2State string,
	token string,
) UserSessionAuthDetails {
	return UserSessionAuthDetails{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "Token",
		},
		OAuth2State: oauth2State,
		Token:       token,
	}
}
