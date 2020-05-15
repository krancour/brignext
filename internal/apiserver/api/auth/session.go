package auth

import (
	"time"

	"github.com/krancour/brignext/v2"
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
