package storage

import "github.com/krancour/brignext/pkg/brignext"

type SessionStore interface {
	CreateSession() (string, string, error)
	CreateRootSession() (string, error)
	GetSessionByOAuth2State(oauth2State string) (*brignext.Session, error)
	GetSessionByToken(token string) (*brignext.Session, error)
	AuthenticateSession(sessionID, userID string) error
	DeleteSession(id string) error
}
