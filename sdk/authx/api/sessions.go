package api

type UserSessionAuthDetails struct {
	OAuth2State string `json:"oauth2State,omitempty"`
	AuthURL     string `json:"authURL,omitempty"`
	Token       string `json:"token,omitempty"`
}
