package oidc

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
)

const envconfigPrefix = "OIDC"

type config struct {
	Enabled bool `envconfig:"ENABLED"`
	// ProviderURL examples:
	//   Google: https://accounts.google.com
	//   Azure Active Directory: https://login.microsoftonline.com/{tenant id}/v2.0
	ProviderURL     string `envconfig:"PROVIDER_URL"`
	ClientID        string `envconfig:"CLIENT_ID"`
	ClientSecret    string `envconfig:"CLIENT_SECRET"`
	RedirectURLBase string `envconfig:"REDIRECT_URL_BASE"`
}

// GetConfigAndVerifierFromEnvironment returns OAuth client configuration and an
// OIDC identity token verifier, all derived from environment variables.
func GetConfigAndVerifierFromEnvironment() (
	*oauth2.Config,
	*oidc.IDTokenVerifier,
	error,
) {
	c := config{}
	if err := envconfig.Process(envconfigPrefix, &c); err != nil {
		return nil, nil, err
	}

	if !c.Enabled {
		return nil, nil, nil // We're not using OIDC
	}

	if c.ProviderURL == "" {
		return nil, nil, errors.New(
			"with OIDC enabled, a value is required for the PROVIDER_URL " +
				"environment variable",
		)
	}
	if c.ClientID == "" {
		return nil, nil, errors.New(
			"with OIDC enabled, a value is required for the CLIENT_ID " +
				"environment variable",
		)
	}
	if c.ClientSecret == "" {
		return nil, nil, errors.New(
			"with OIDC enabled, a value is required for the CLIENT_SECRET " +
				"environment variable",
		)
	}
	if c.RedirectURLBase == "" {
		return nil, nil, errors.New(
			"with OIDC enabled, a value is required for the REDIRECT_URL_BASE " +
				"environment variable",
		)
	}

	provider, err := oidc.NewProvider(context.TODO(), c.ProviderURL)
	if err != nil {
		return nil, nil, err
	}

	config := &oauth2.Config{
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

	return config, verifier, nil
}
