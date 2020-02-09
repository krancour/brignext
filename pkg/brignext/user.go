package brignext

import (
	"time"
)

type User struct {
	ID        string    `json:"id,omitempty" bson:"_id,omitempty"`
	Name      string    `json:"name,omitempty" bson:"name,omitempty"`
	FirstSeen time.Time `json:"firstSeen,omitempty" bson:"firstSeen,omitempty"`
}
