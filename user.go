package brignext

import (
	"time"
)

type User struct {
	ID        string    `json:"id,omitempty" bson:"_id"`
	Name      string    `json:"name,omitempty" bson:"name"`
	FirstSeen time.Time `json:"firstSeen,omitempty" bson:"firstSeen"`
	Locked    *bool     `json:"locked,omitempty" bson:"locked"`
}
