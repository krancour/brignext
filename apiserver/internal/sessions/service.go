package sessions

import (
	"context"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/krancour/brignext/v2/internal/api/auth"
	"github.com/krancour/brignext/v2/internal/crypto"
	brignext "github.com/krancour/brignext/v2/sdk"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type Service interface {
	CreateRootSession(context.Context) (brignext.Token, error)
	CreateUserSession(context.Context) (brignext.UserSessionAuthDetails, error)
	GetByOAuth2State(context.Context, string) (auth.Session, error)
	GetByToken(context.Context, string) (auth.Session, error)
	Authenticate(
		ctx context.Context,
		sessionID string,
		userID string,
	) error
	Delete(context.Context, string) error

	CheckHealth(context.Context) error
}

type service struct {
	store             Store
	oauth2Config      *oauth2.Config
	oidcTokenVerifier *oidc.IDTokenVerifier
}

func NewService(
	store Store,
	oauth2Config *oauth2.Config,
	oidcTokenVerifier *oidc.IDTokenVerifier,
) Service {
	return &service{
		store:             store,
		oauth2Config:      oauth2Config,
		oidcTokenVerifier: oidcTokenVerifier,
	}
}

func (s *service) CreateRootSession(
	ctx context.Context,
) (brignext.Token, error) {
	token := brignext.NewToken(crypto.NewToken(256))
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
	return userSessionAuthDetails, nil
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

func (s *service) Authenticate(
	ctx context.Context,
	sessionID string,
	userID string,
) error {
	if err := s.store.Authenticate(
		ctx,
		sessionID,
		userID,
		time.Now().Add(time.Hour),
	); err != nil {
		return errors.Wrapf(
			err,
			"error storing authentication details for session %q",
			sessionID,
		)
	}
	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	if err := s.store.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error removing session %q from store", id)
	}
	return nil
}

func (s *service) CheckHealth(ctx context.Context) error {
	if err := s.store.CheckHealth(ctx); err != nil {
		return errors.Wrap(err, "error checking sessions store health")
	}
	return nil
}
