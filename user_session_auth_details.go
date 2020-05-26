package brignext

type UserSessionAuthDetails struct {
	TypeMeta    `json:",inline" bson:",inline"`
	OAuth2State string `json:"oauth2State"`
	AuthURL     string `json:"authURL"`
	Token       string `json:"token"`
}

func NewUserSessionAuthDetails(
	oauth2State string,
	token string,
) UserSessionAuthDetails {
	return UserSessionAuthDetails{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "Token",
		},
		OAuth2State: oauth2State,
		Token:       token,
	}
}
