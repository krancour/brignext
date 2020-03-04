package brignext

import (
	"time"
)

type User struct {
	ID        string    `json:"id" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	FirstSeen time.Time `json:"firstSeen" bson:"firstSeen"`
	Locked    bool      `json:"locked" bson:"locked"`
}
