package storage

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2/internal/apiserver/api/auth"
)

type SessionsStore interface {
	Create(context.Context, auth.Session) error
	GetByHashedOAuth2State(context.Context, string) (auth.Session, error)
	GetByHashedToken(context.Context, string) (auth.Session, error)
	Authenticate(
		ctx context.Context,
		sessionID string,
		userID string,
		expires time.Time,
	) error
	Delete(context.Context, string) error
	DeleteByUser(context.Context, string) (int64, error)
}
