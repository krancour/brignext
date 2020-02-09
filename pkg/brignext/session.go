package brignext

import (
	"time"
)

type Session struct {
	ID            string    `json:"id,omitempty" bson:"_id,omitempty"`
	OAuth2State   string    `json:"-" bson:"-"`
	Token         string    `json:"-" bson:"-"`
	Root          bool      `json:"root,omitempty" bson:"root,omitempty"`
	Authenticated bool      `json:"authenticated,omitempty" bson:"authenticated,omitempty"`
	UserID        string    `json:"userID,omitempty" bson:"userID,omitempty"`
	Expires       time.Time `json:"expires,omitempty" bson:"expires,omitempty"`
}
