package authx

import (
	"context"
	"encoding/json"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/oauth2"
)

// OIDCAuthDetails encapsulates all information required for a client
// authenticating by means of OpenID Connect to complete the authentication
// process using a third-party identiy provider.
type OIDCAuthDetails struct {
	// OAuth2State is an opaque token issued by Brigade that must be sent to the
	// OIDC identity provider as a query parameter when the OIDC authentication
	// workflow continues (in the user's web browser). The OIDC identity provider
	// includes this token when it issues a callback to the Brigade API server
	// after successful authentication. This permits the Brigade API server to
	// correlate the successful authentication by the OIDC identity provider with
	// an existing, but as-yet-unactivated token. This proof of authentication
	// allows Brigade to activate the token and associate it with the User that
	// the OIDC identity provider indicates has successfully completed
	// authentication.
	OAuth2State string `json:"oauth2State"`
	// AuthURL is a URL that can be requested in a user's web browser to complete
	// authentication via a third-party OIDC identity provider.
	AuthURL string `json:"authURL"`
	// Token is an opaque bearer token issued by Brigade to correlate a User with
	// a Session. It remains unactivated (useless) until the OIDC authentication
	// workflow is successfully completed. Clients may expect that that the token
	// expires (at an interval determined by a system administrator) and, for
	// simplicity, is NOT refreshable. When the token has expired, it
	// re-authentication is required.
	Token string `json:"token"`
}

// MarshalJSON amends OIDCAuthDetails instances with type metadata.
func (o OIDCAuthDetails) MarshalJSON() ([]byte, error) {
	type Alias OIDCAuthDetails
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "OIDCAuthDetails",
			},
			Alias: (Alias)(o),
		},
	)
}

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

// SessionsService is the specialized interface for managing Sessions. It's
// decoupled from underlying technology choices (e.g. data store) to keep
// business logic reusable and consistent while the underlying tech stack
// remains free to change.
type SessionsService interface {
	// CreateRootSession creates a Session for the root user (if enabled by th
	// system administrator) and returns a Token with a short expiry period
	// (determined by a system administrator). If the specified username is not
	// "root" or the specified password is incorrect, implementations MUST return
	// a *meta.ErrAuthentication error.
	CreateRootSession(
		ctx context.Context,
		username string,
		password string,
	) (Token, error)
	// CreateUserSession creates a new User Session and initiates an OpenID
	// Connect authentication workflow. It returns an OIDCAuthDetails containing
	// all information required to continue the authentication process with a
	// third-party OIDC identity provider.
	CreateUserSession(context.Context) (OIDCAuthDetails, error)
	// Authenticate completes the final steps of the OpenID Connect authentication
	// workflow. It uses the provided oauth2State to idenity an as-yet anonymous
	// Session (with) an as-yet unactivated token). It communicates with the
	// third-party ODIC identity provider, exchanging the provided oidcCode for
	// user information. This information can be used to correlate the as-yet
	// anonymous Session to an existing User. If the User is previously unknown to
	// Brigade, one is seamlessly created (with read-only permissions) form the
	// information provided by the identity provider. Finally, the Session's token
	// is activated.
	Authenticate(
		ctx context.Context,
		oauth2State string,
		oidcCode string,
	) error
	// GetByToken retrieves the Session having the provided token. If no such
	// Session is found or is found but is expired, implementations MUST return a
	// *meta.ErrAuthentication error.
	GetByToken(ctx context.Context, token string) (Session, error)
	// Delete deletes the specified Session.
	Delete(ctx context.Context, id string) error
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
) (OIDCAuthDetails, error) {
	userSessionAuthDetails := OIDCAuthDetails{
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
