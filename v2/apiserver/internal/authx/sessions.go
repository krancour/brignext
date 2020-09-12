package authx

import (
	"context"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/oauth2"
)

type Session struct {
	meta.TypeMeta     `json:",inline" bson:",inline"`
	meta.ObjectMeta   `json:"metadata" bson:",inline"`
	Root              bool       `json:"root" bson:"root"`
	UserID            string     `json:"userID" bson:"userID"`
	HashedOAuth2State string     `json:"-" bson:"hashedOAuth2State"`
	HashedToken       string     `json:"-" bson:"hashedToken"`
	Authenticated     *time.Time `json:"authenticated" bson:"authenticated"`
	Expires           *time.Time `json:"expires" bson:"expires"`
}

func NewRootSession(token string) Session {
	now := time.Now()
	expiryTime := now.Add(time.Hour)
	return Session{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "Session",
		},
		ObjectMeta: meta.ObjectMeta{
			ID: uuid.NewV4().String(),
		},
		Root:          true,
		HashedToken:   crypto.ShortSHA("", token),
		Authenticated: &now,
		Expires:       &expiryTime,
	}
}

func NewUserSession(oauth2State, token string) Session {
	return Session{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "Session",
		},
		ObjectMeta: meta.ObjectMeta{
			ID: uuid.NewV4().String(),
		},
		HashedOAuth2State: crypto.ShortSHA("", oauth2State),
		HashedToken:       crypto.ShortSHA("", token),
	}
}

type sessionIDContextKey struct{}

func ContextWithSessionID(
	ctx context.Context,
	sessionID string,
) context.Context {
	return context.WithValue(ctx, sessionIDContextKey{}, sessionID)
}

func SessionIDFromContext(ctx context.Context) string {
	token := ctx.Value(sessionIDContextKey{})
	if token == nil {
		return ""
	}
	return token.(string)
}

// TODO: We probably don't need this interface. The idea is to have a single
// implementation of the service's logic, with only underlying components being
// pluggable. BUT, STRONGLY CONSIDER THAT WE MAY NEED THIS TO MOCK OUT THE
// SERVICE WHEN TESTING THE CORRESPONDING ENDPOINTS.
type SessionsService interface {
	CreateRootSession(
		ctx context.Context,
		username string,
		password string,
	) (Token, error)
	CreateUserSession(context.Context) (UserSessionAuthDetails, error)
	Authenticate(
		ctx context.Context,
		oauth2State string,
		oidcCode string,
	) error
	GetByToken(context.Context, string) (Session, error)
	Delete(context.Context, string) error
}

type sessionsService struct {
	authorize              AuthorizeFn
	sessionsStore          SessionsStore
	usersStore             UsersStore
	rootUserEnabled        bool
	hashedRootUserPassword string
	oauth2Config           *oauth2.Config
	oidcTokenVerifier      *oidc.IDTokenVerifier
}

// TODO: There probably isn't any good reason to actually have this
// constructor-like function here. Let's consider removing it.
func NewSessionsService(
	sessionsStore SessionsStore,
	usersStore UsersStore,
	rootUserEnabled bool,
	hashedRootUserPassword string,
	oauth2Config *oauth2.Config,
	oidcTokenVerifier *oidc.IDTokenVerifier,
) SessionsService {
	return &sessionsService{
		authorize:              Authorize,
		sessionsStore:          sessionsStore,
		usersStore:             usersStore,
		rootUserEnabled:        rootUserEnabled,
		hashedRootUserPassword: hashedRootUserPassword,
		oauth2Config:           oauth2Config,
		oidcTokenVerifier:      oidcTokenVerifier,
	}
}

func (s *sessionsService) CreateRootSession(
	ctx context.Context,
	username string,
	password string,
) (Token, error) {
	token := Token{
		Value: crypto.NewToken(256),
	}
	if !s.rootUserEnabled {
		return token, &meta.ErrNotSupported{
			Details: "Authentication using root credentials is not supported by " +
				"this server.",
		}
	}
	if username != "root" ||
		crypto.ShortSHA(username, password) != s.hashedRootUserPassword {
		return token, &meta.ErrAuthentication{
			Reason: "Could not authenticate request using the supplied credentials.",
		}
	}
	session := NewRootSession(token.Value)
	now := time.Now()
	session.Created = &now
	if err := s.sessionsStore.Create(ctx, session); err != nil {
		return token, errors.Wrapf(
			err,
			"error storing new root session %q",
			session.ID,
		)
	}
	return token, nil
}

func (s *sessionsService) CreateUserSession(
	ctx context.Context,
) (UserSessionAuthDetails, error) {
	userSessionAuthDetails := UserSessionAuthDetails{
		OAuth2State: crypto.NewToken(30),
		Token:       crypto.NewToken(256),
	}
	session := NewUserSession(
		userSessionAuthDetails.OAuth2State,
		userSessionAuthDetails.Token,
	)
	now := time.Now()
	session.Created = &now
	if err := s.sessionsStore.Create(ctx, session); err != nil {
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

func (s *sessionsService) Authenticate(
	ctx context.Context,
	oauth2State string,
	oidcCode string,
) error {
	if s.oauth2Config == nil || s.oidcTokenVerifier == nil {
		return &meta.ErrNotSupported{
			Details: "Authentication using OpenID Connect is not supported by this " +
				"server.",
		}
	}
	session, err := s.sessionsStore.GetByHashedOAuth2State(
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
		if _, ok := errors.Cause(err).(*meta.ErrNotFound); ok {
			// User wasn't found. That's ok. We'll create one.
			user = User{
				ObjectMeta: meta.ObjectMeta{
					ID: claims.Email,
				},
				Name: claims.Name,
			}

			// User 0 gets a bunch of roles automatically
			var count int64
			count, err = s.usersStore.Count(ctx)
			if err != nil {
				return errors.Wrap(err, "error counting users in store")
			}
			if count == 0 {
				user.UserRoles = []Role{
					RoleAdmin(),
					RoleProjectCreator(),
					RoleReader(),
				}
			}

			if err = s.usersStore.Create(ctx, user); err != nil {
				return errors.Wrapf(err, "error storing new user %q", user.ID)
			}
		} else {
			// It was something else that went wrong when searching for the user.
			return err
		}
	}
	if err := s.sessionsStore.Authenticate(
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

func (s *sessionsService) GetByOAuth2State(
	ctx context.Context,
	oauth2State string,
) (Session, error) {
	session, err := s.sessionsStore.GetByHashedOAuth2State(
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

func (s *sessionsService) GetByToken(
	ctx context.Context,
	token string,
) (Session, error) {
	session, err := s.sessionsStore.GetByHashedToken(
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

func (s *sessionsService) Delete(ctx context.Context, id string) error {
	if err := s.sessionsStore.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error removing session %q from store", id)
	}
	return nil
}

type SessionsStore interface {
	Create(context.Context, Session) error
	GetByHashedOAuth2State(context.Context, string) (Session, error)
	GetByHashedToken(context.Context, string) (Session, error)
	Authenticate(
		ctx context.Context,
		sessionID string,
		userID string,
		expires time.Time,
	) error
	Delete(context.Context, string) error
}
