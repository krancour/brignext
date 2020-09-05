package authx

import (
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	uuid "github.com/satori/go.uuid"
)

type Session struct {
	meta.TypeMeta     `json:",inline" bson:",inline"`
	meta.ObjectMeta   `json:"metadata" bson:",inline"`
	Root              bool       `json:"root" bson:"root"`
	UserID            string     `json:"userID" bson:"userID"`
	HashedOAuth2State string     `json:"-" bson:"hashedOAuth2State"`
	HashedToken       string     `json:"-" bson:"hashedToken"`
	Authenticated     *time.Time `json:"authenticated" bson:"authenticated"`
	Expires           *time.Time `json:"expires" bson:"expires"`
}

func NewRootSession(token string) Session {
	now := time.Now()
	expiryTime := now.Add(time.Hour)
	return Session{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "Session",
		},
		ObjectMeta: meta.ObjectMeta{
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
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "Session",
		},
		ObjectMeta: meta.ObjectMeta{
			ID: uuid.NewV4().String(),
		},
		HashedOAuth2State: crypto.ShortSHA("", oauth2State),
		HashedToken:       crypto.ShortSHA("", token),
	}
}
