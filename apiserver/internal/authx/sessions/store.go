package sessions

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
)

type Store interface {
	Create(context.Context, authx.Session) error
	GetByHashedOAuth2State(context.Context, string) (authx.Session, error)
	GetByHashedToken(context.Context, string) (authx.Session, error)
	Authenticate(
		ctx context.Context,
		sessionID string,
		userID string,
		expires time.Time,
	) error
	Delete(context.Context, string) error
}
