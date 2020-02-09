package brignext

import (
	"time"
)

type User struct {
	ID        string    `json:"id" bson:"_id"`
	Username  string    `json:"username" bson:"username"`
	Name      string    `json:"name" bson:"name"`
	FirstSeen time.Time `json:"firstSeen" bson:"firstSeen"`
}
