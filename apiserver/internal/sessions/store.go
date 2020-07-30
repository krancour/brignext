package sessions

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery/auth"
)

type Store interface {
	// TODO: This looks a little off. Sessions aren't part of the SDK, but they
	// should be split off into some other package that isn't under API machinery.
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
}
