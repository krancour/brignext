package brignext

import (
	"time"
)

type Session struct {
	ID                string    `json:"-" bson:"_id"`
	Root              bool      `json:"-" bson:"root"`
	Authenticated     bool      `json:"-" bson:"authenticated"`
	UserID            string    `json:"-" bson:"userID"`
	HashedOAuth2State string    `json:"-" bson:"hashedOAuth2State"`
	HashedToken       string    `json:"-" bson:"hashedToken"`
	Created           time.Time `json:"-" bson:"created"`
	Expires           time.Time `json:"-" bson:"expires"`
}
