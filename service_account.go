package brignext

import "time"

type ServiceAccountList struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata"`
	Items    []ServiceAccount `json:"items"`
}

type ServiceAccount struct {
	TypeMeta    `json:",inline" bson:",inline"`
	ObjectMeta  `json:"metadata" bson:"metadata"`
	Description string     `json:"description" bson:"description"`
	HashedToken string     `json:"-" bson:"hashedToken"`
	Locked      *time.Time `json:"locked,omitempty" bson:"locked"`
}
