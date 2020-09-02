package sessions

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/authn"
)

type Store interface {
	Create(context.Context, authn.Session) error
	GetByHashedOAuth2State(context.Context, string) (authn.Session, error)
	GetByHashedToken(context.Context, string) (authn.Session, error)
	Authenticate(
		ctx context.Context,
		sessionID string,
		userID string,
		expires time.Time,
	) error
	Delete(context.Context, string) error
}
