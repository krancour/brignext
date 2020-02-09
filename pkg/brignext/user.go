package brignext

import (
	"time"
)

type User struct {
	ID        string    `json:"-" bson:"_id,omitempty"`
	Username  string    `json:"username,omitempty" bson:"username,omitempty"`
	Name      string    `json:"name,omitempty" bson:"name,omitempty"`
	FirstSeen time.Time `json:"firstSeen,omitempty" bson:"firstSeen,omitempty"`
}
