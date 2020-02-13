package brignext

import (
	"time"
)

type Session struct {
	ID                string    `json:"-" bson:"_id,omitempty"`
	Root              bool      `json:"-" bson:"root,omitempty"`
	Authenticated     bool      `json:"-" bson:"authenticated,omitempty"`
	UserID            string    `json:"-" bson:"userID,omitempty"`
	HashedOAuth2State string    `json:"-" bson:"hashedOAuth2State,omitempty"`
	HashedToken       string    `json:"-" bson:"hashedToken,omitempty"`
	Created           time.Time `json:"-" bson:"created,omitempty"`
	Expires           time.Time `json:"-" bson:"expires,omitempty"`
}
