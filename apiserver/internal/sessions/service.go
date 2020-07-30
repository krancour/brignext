package sessions

import (
	"context"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/krancour/brignext/v2/apiserver/internal/api/auth"
	"github.com/krancour/brignext/v2/apiserver/internal/crypto"
	"github.com/krancour/brignext/v2/apiserver/internal/users"
	errs "github.com/krancour/brignext/v2/internal/errors"
	brignext "github.com/krancour/brignext/v2/sdk"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type Service interface {
	CreateRootSession(
		ctx context.Context,
		username string,
		password string,
	) (brignext.Token, error)
	CreateUserSession(context.Context) (brignext.UserSessionAuthDetails, error)
	Authenticate(
		ctx context.Context,
		oauth2State string,
		oidcCode string,
	) error
	GetByToken(context.Context, string) (auth.Session, error)
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
) (brignext.Token, error) {
	token := brignext.NewToken(crypto.NewToken(256))
	if !s.rootUserEnabled {
		return token, errs.NewErrNotSupported(
			"Authentication using root credentials is not supported by " +
				"this server.",
		)
	}
	if username != "root" ||
		crypto.ShortSHA(username, password) != s.hashedRootUserPassword {
		return token, errs.NewErrAuthentication(
			"Could not authenticate request using the supplied credentials.",
		)
	}
	session := auth.NewRootSession(token.Value)
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
) (brignext.UserSessionAuthDetails, error) {
	userSessionAuthDetails := brignext.NewUserSessionAuthDetails(
		crypto.NewToken(30),
		crypto.NewToken(256),
	)
	session := auth.NewUserSession(
		userSessionAuthDetails.OAuth2State,
		userSessionAuthDetails.Token,
	)
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
		return errs.NewErrNotSupported(
			"Authentication using OpenID Connect is not supported by this " +
				"server.",
		)
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
		if _, ok := errors.Cause(err).(*errs.ErrNotFound); ok {
			// User wasn't found. That's ok. We'll create one.
			user = brignext.NewUser(claims.Email, claims.Name)
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
) (auth.Session, error) {
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
) (auth.Session, error) {
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
