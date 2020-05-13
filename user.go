package brignext

import (
	"time"
)

type User struct {
	TypeMeta `json:",inline" bson:",inline"`
	UserMeta `json:"metadata" bson:"metadata"`
	Status   UserStatus `json:"status,omitempty" bson:"status"`
}

type UserMeta struct {
	ID        string    `json:"id" bson:"id"`
	Name      string    `json:"name" bson:"name"`
	FirstSeen time.Time `json:"firstSeen" bson:"firstSeen"`
}

type UserStatus struct {
	Locked bool `json:"locked" bson:"locked"`
}
