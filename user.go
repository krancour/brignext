package brignext

import "time"

type User struct {
	TypeMeta   `json:",inline" bson:",inline"`
	ObjectMeta `json:"metadata" bson:"metadata"`
	Name       string     `json:"name" bson:"name"`
	Locked     *time.Time `json:"locked" bson:"locked"`
}
