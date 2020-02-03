package oidc

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
)

const envconfigPrefix = "OIDC"

type config struct {
	// ProviderURL examples:
	//   Google: https://accounts.google.com
	//   Azure Active Directory: https://login.microsoftonline.com/{tenant id}/v2.0
	ProviderURL     string `envconfig:"PROVIDER_URL" required:"true"`
	ClientID        string `envconfig:"CLIENT_ID" required:"true"`
	ClientSecret    string `envconfig:"CLIENT_SECRET" required:"true"`
	RedirectURLBase string `envconfig:"REDIRECT_URL_BASE" required:"true"`
}

// GetConfigAndVerifierFromEnvironment returns OAuth client configuration and an
// OIDC identity token verifier, all derived from environment variables.
func GetConfigAndVerifierFromEnvironment() (
	oauth2.Config,
	oidc.IDTokenVerifier,
	error,
) {
	c := config{}
	if err := envconfig.Process(envconfigPrefix, &c); err != nil {
		return oauth2.Config{}, oidc.IDTokenVerifier{}, err
	}

	provider, err := oidc.NewProvider(context.TODO(), c.ProviderURL)
	if err != nil {
		return oauth2.Config{}, oidc.IDTokenVerifier{}, err
	}

	config := oauth2.Config{
		Endpoint:     provider.Endpoint(),
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL: fmt.Sprintf(
			"%s/%s",
			c.RedirectURLBase,
			"auth/oidc/callback",
		),
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(
		&oidc.Config{
			ClientID: c.ClientID,
		},
	)

	return config, *verifier, nil
}
