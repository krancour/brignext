package brignext

import (
	"time"
)

type Session struct {
	ID            string    `json:"id,omitempty" bson:"_id,omitempty"`
	Root          bool      `json:"root,omitempty" bson:"root,omitempty"`
	Authenticated bool      `json:"authenticated,omitempty" bson:"authenticated,omitempty"`
	UserID        string    `json:"userID,omitempty" bson:"userID,omitempty"`
	Created       time.Time `json:"created,omitempty" bson:"created,omitempty"`
	Expires       time.Time `json:"expires,omitempty" bson:"expires,omitempty"`
}
