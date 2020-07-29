package sessions

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/api/auth"
)

type Store interface {
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

	DoTx(context.Context, func(context.Context) error) error

	CheckHealth(context.Context) error
}
