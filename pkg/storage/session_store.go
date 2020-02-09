package storage

import "github.com/krancour/brignext/pkg/brignext"

type SessionStore interface {
	CreateSession(session brignext.Session) (string, string, string, error)
	GetSessionByOAuth2State(oauth2State string) (brignext.Session, bool, error)
	GetSessionByToken(token string) (brignext.Session, bool, error)
	AuthenticateSession(sessionID, userID string) error
	DeleteSession(id string) error
}
