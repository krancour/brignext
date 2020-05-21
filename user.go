package brignext

import "time"

type UserList struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata"`
	Items    []User `json:"items"`
}

func NewUserList() UserList {
	return UserList{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "UserList",
		},
		Items: []User{},
	}
}

type User struct {
	TypeMeta   `json:",inline" bson:",inline"`
	ObjectMeta `json:"metadata" bson:"metadata"`
	Name       string     `json:"name" bson:"name"`
	Locked     *time.Time `json:"locked" bson:"locked"`
}

func NewUser(id, name string) User {
	return User{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "User",
		},
		ObjectMeta: ObjectMeta{
			ID: id,
		},
		Name: name,
	}
}
