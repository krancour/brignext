package authx

import (
	"context"
	"time"
)

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
