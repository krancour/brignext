package sessions

import (
	"context"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/krancour/brignext/v2/apiserver/internal/authn"
	"github.com/krancour/brignext/v2/apiserver/internal/authn/users"
	"github.com/krancour/brignext/v2/apiserver/internal/crypto"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type Service interface {
	CreateRootSession(
		ctx context.Context,
		username string,
		password string,
	) (authn.Token, error)
	CreateUserSession(context.Context) (authn.UserSessionAuthDetails, error)
	Authenticate(
		ctx context.Context,
		oauth2State string,
		oidcCode string,
	) error
	GetByToken(context.Context, string) (authn.Session, error)
	Delete(context.Context, string) error
}

type service struct {
	store                  Store
	usersStore             users.Store
	rootUserEnabled        bool
	hashedRootUserPassword string
	oauth2Config           *oauth2.Config
	oidcTokenVerifier      *oidc.IDTokenVerifier
}

func NewService(
	store Store,
	usersStore users.Store,
	rootUserEnabled bool,
	hashedRootUserPassword string,
	oauth2Config *oauth2.Config,
	oidcTokenVerifier *oidc.IDTokenVerifier,
) Service {
	return &service{
		store:                  store,
		usersStore:             usersStore,
		rootUserEnabled:        rootUserEnabled,
		hashedRootUserPassword: hashedRootUserPassword,
		oauth2Config:           oauth2Config,
		oidcTokenVerifier:      oidcTokenVerifier,
	}
}

func (s *service) CreateRootSession(
	ctx context.Context,
	username string,
	password string,
) (authn.Token, error) {
	token := authn.Token{
		Value: crypto.NewToken(256),
	}
	if !s.rootUserEnabled {
		return token, &brignext.ErrNotSupported{
			Details: "Authentication using root credentials is not supported by " +
				"this server.",
		}
	}
	if username != "root" ||
		crypto.ShortSHA(username, password) != s.hashedRootUserPassword {
		return token, &brignext.ErrAuthentication{
			Reason: "Could not authenticate request using the supplied credentials.",
		}
	}
	session := authn.NewRootSession(token.Value)
	now := time.Now()
	session.Created = &now
	if err := s.store.Create(ctx, session); err != nil {
		return token, errors.Wrapf(
			err,
			"error storing new root session %q",
			session.ID,
		)
	}
	return token, nil
}

func (s *service) CreateUserSession(
	ctx context.Context,
) (authn.UserSessionAuthDetails, error) {
	userSessionAuthDetails := authn.UserSessionAuthDetails{
		OAuth2State: crypto.NewToken(30),
		Token:       crypto.NewToken(256),
	}
	session := authn.NewUserSession(
		userSessionAuthDetails.OAuth2State,
		userSessionAuthDetails.Token,
	)
	now := time.Now()
	session.Created = &now
	if err := s.store.Create(ctx, session); err != nil {
		return userSessionAuthDetails, errors.Wrapf(
			err,
			"error storing new user session %q",
			session.ID,
		)
	}
	userSessionAuthDetails.AuthURL = s.oauth2Config.AuthCodeURL(
		userSessionAuthDetails.OAuth2State,
	)
	return userSessionAuthDetails, nil
}

func (s *service) Authenticate(
	ctx context.Context,
	oauth2State string,
	oidcCode string,
) error {
	if s.oauth2Config == nil || s.oidcTokenVerifier == nil {
		return &brignext.ErrNotSupported{
			Details: "Authentication using OpenID Connect is not supported by this " +
				"server.",
		}
	}
	session, err := s.store.GetByHashedOAuth2State(
		ctx,
		crypto.ShortSHA("", oauth2State),
	)
	if err != nil {
		return errors.Wrap(
			err,
			"error retrieving session from store by hashed OAuth2 state",
		)
	}
	oauth2Token, err := s.oauth2Config.Exchange(ctx, oidcCode)
	if err != nil {
		return errors.Wrap(
			err,
			"error exchanging OpenID Connect code for OAuth2 token",
		)
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return errors.New(
			"OAuth2 token, did not include an OpenID Connect identity token",
		)
	}
	idToken, err := s.oidcTokenVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		return errors.Wrap(err, "error verifying OpenID Connect identity token")
	}
	claims := struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}{}
	if err = idToken.Claims(&claims); err != nil {
		return errors.Wrap(
			err,
			"error decoding OpenID Connect identity token claims",
		)
	}
	user, err := s.usersStore.Get(ctx, claims.Email)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrNotFound); ok {
			// User wasn't found. That's ok. We'll create one.
			user = authn.User{
				ObjectMeta: meta.ObjectMeta{
					ID: claims.Email,
				},
				Name: claims.Name,
			}
			if err = s.usersStore.Create(ctx, user); err != nil {
				return err
			}
		} else {
			// It was something else that went wrong when searching for the user.
			return err
		}
	}
	if err := s.store.Authenticate(
		ctx,
		session.ID,
		user.ID,
		time.Now().Add(time.Hour),
	); err != nil {
		return errors.Wrapf(
			err,
			"error storing authentication details for session %q",
			session.ID,
		)
	}
	return nil
}

func (s *service) GetByOAuth2State(
	ctx context.Context,
	oauth2State string,
) (authn.Session, error) {
	session, err := s.store.GetByHashedOAuth2State(
		ctx,
		crypto.ShortSHA("", oauth2State),
	)
	if err != nil {
		return session, errors.Wrap(
			err,
			"error retrieving session from store by hashed oauth2 state",
		)
	}
	return session, nil
}

func (s *service) GetByToken(
	ctx context.Context,
	token string,
) (authn.Session, error) {
	session, err := s.store.GetByHashedToken(
		ctx,
		crypto.ShortSHA("", token),
	)
	if err != nil {
		return session, errors.Wrap(
			err,
			"error retrieving session from store by hashed token",
		)
	}
	return session, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	if err := s.store.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error removing session %q from store", id)
	}
	return nil
}
