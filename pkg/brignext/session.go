package brignext

import (
	"time"
)

type Session struct {
	ID            string    `json:"id" bson:"id"`
	OAuth2State   string    `json:"-" bson:"-"`
	Token         string    `json:"-" bson:"-"`
	Root          bool      `json:"root" bson:"root"`
	Authenticated bool      `json:"authenticated" bson:"authenticated"`
	Username      string    `json:"username" bson:"username,omitempty"`
	Expires       time.Time `json:"expires" bson:"expires"`
}
