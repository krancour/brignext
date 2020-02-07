package brignext

import (
	"time"
)

type User struct {
	Username  string    `json:"username" bson:"username"`
	Name      string    `json:"name" bson:"name"`
	FirstSeen time.Time `json:"firstSeen" bson:"firstSeen"`
}
