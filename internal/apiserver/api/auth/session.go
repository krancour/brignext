package auth

import (
	"time"

	"github.com/krancour/brignext/v2"
)

type Session struct {
	brignext.TypeMeta `json:",inline" bson:",inline"`
	SessionMeta       `json:"metadata" bson:"metadata"`
	Spec              SessionSpec   `json:"spec" bson:"spec"`
	Status            SessionStatus `json:"status" bson:"status"`
}

type SessionMeta struct {
	ID      string    `json:"id" bson:"id"`
	Created time.Time `json:"created" bson:"created"`
	Expires time.Time `json:"expires" bson:"expires"`
}

type SessionSpec struct {
	Root              bool   `json:"root" bson:"root"`
	UserID            string `json:"userID" bson:"userID"`
	HashedOAuth2State string `json:"-" bson:"hashedOAuth2State"`
	HashedToken       string `json:"-" bson:"hashedToken"`
}

type SessionStatus struct {
	Authenticated bool `json:"authenticated" bson:"authenticated"`
}
