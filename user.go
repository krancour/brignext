package brignext

import "time"

type UserList struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata"`
	Items    []User `json:"items"`
}

type User struct {
	TypeMeta   `json:",inline" bson:",inline"`
	ObjectMeta `json:"metadata" bson:"metadata"`
	Name       string     `json:"name" bson:"name"`
	Locked     *time.Time `json:"locked" bson:"locked"`
}
