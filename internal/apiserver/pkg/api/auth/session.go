package auth

import (
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/pkg/crypto"
	uuid "github.com/satori/go.uuid"
)

type Session struct {
	brignext.TypeMeta   `json:",inline" bson:",inline"`
	brignext.ObjectMeta `json:"metadata" bson:"metadata"`
	Root                bool       `json:"root" bson:"root"`
	UserID              string     `json:"userID" bson:"userID"`
	HashedOAuth2State   string     `json:"-" bson:"hashedOAuth2State"`
	HashedToken         string     `json:"-" bson:"hashedToken"`
	Authenticated       *time.Time `json:"authenticated" bson:"authenticated"`
	Expires             *time.Time `json:"expires" bson:"expires"`
}

func NewRootSession(token string) Session {
	now := time.Now()
	expiryTime := now.Add(time.Hour)
	return Session{
		TypeMeta: brignext.TypeMeta{
			APIVersion: brignext.APIVersion,
			Kind:       "Session",
		},
		ObjectMeta: brignext.ObjectMeta{
			ID: uuid.NewV4().String(),
		},
		Root:          true,
		HashedToken:   crypto.ShortSHA("", token),
		Authenticated: &now,
		Expires:       &expiryTime,
	}
}

func NewUserSession(oauth2State, token string) Session {
	return Session{
		TypeMeta: brignext.TypeMeta{
			APIVersion: brignext.APIVersion,
			Kind:       "Session",
		},
		ObjectMeta: brignext.ObjectMeta{
			ID: uuid.NewV4().String(),
		},
		HashedOAuth2State: crypto.ShortSHA("", oauth2State),
		HashedToken:       crypto.ShortSHA("", token),
	}
}
