package service

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/api/auth"
	"github.com/krancour/brignext/v2/internal/apiserver/crypto"
	"github.com/krancour/brignext/v2/internal/apiserver/storage"
	"github.com/pkg/errors"
)

type SessionsService interface {
	CreateRootSession(context.Context) (brignext.Token, error)
	CreateUserSession(context.Context) (string, string, error)
	GetByOAuth2State(context.Context, string) (auth.Session, error)
	GetByToken(context.Context, string) (auth.Session, error)
	Authenticate(
		ctx context.Context,
		sessionID string,
		userID string,
	) error
	Delete(context.Context, string) error
	DeleteByUser(context.Context, string) (int64, error)
}

type sessionsService struct {
	store storage.SessionsStore
}

func NewSessionsService(store storage.SessionsStore) SessionsService {
	return &sessionsService{
		store: store,
	}
}

func (s *sessionsService) CreateRootSession(
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

func (s *sessionsService) CreateUserSession(
	ctx context.Context,
) (string, string, error) {
	oauth2State := crypto.NewToken(30)
	token := crypto.NewToken(256)
	session := auth.NewUserSession(oauth2State, token)
	if err := s.store.Create(ctx, session); err != nil {
		return "", "", errors.Wrapf(
			err,
			"error storing new user session %q",
			session.ID,
		)
	}
	return oauth2State, token, nil
}

func (s *sessionsService) GetByOAuth2State(
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

func (s *sessionsService) GetByToken(
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

func (s *sessionsService) Authenticate(
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

func (s *sessionsService) Delete(ctx context.Context, id string) error {
	if err := s.store.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error removing session %q from store", id)
	}
	return nil
}

func (s *sessionsService) DeleteByUser(
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
