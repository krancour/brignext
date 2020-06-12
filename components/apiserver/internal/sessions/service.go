package sessions

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/api/auth"
	"github.com/krancour/brignext/v2/internal/crypto"
	"github.com/pkg/errors"
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
	DeleteByUser(context.Context, string) (int64, error)

	CheckHealth(context.Context) error
}

type service struct {
	store Store
}

func NewService(store Store) Service {
	return &service{
		store: store,
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

func (s *service) DeleteByUser(
	ctx context.Context,
	userID string,
) (int64, error) {
	n, err := s.store.DeleteByUser(ctx, userID)
	if err != nil {
		return 0, errors.Wrapf(
			err,
			"error removing sessions for user %q from store",
			userID,
		)
	}
	return n, nil
}

func (s *service) CheckHealth(ctx context.Context) error {
	return s.store.CheckHealth(ctx)
}
