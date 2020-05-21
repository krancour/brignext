package brignext

import "time"

type ServiceAccountList struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata"`
	Items    []ServiceAccount `json:"items"`
}

func NewServiceAccountList() ServiceAccountList {
	return ServiceAccountList{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "ServiceAccountList",
		},
		Items: []ServiceAccount{},
	}
}

type ServiceAccount struct {
	TypeMeta    `json:",inline" bson:",inline"`
	ObjectMeta  `json:"metadata" bson:"metadata"`
	Description string     `json:"description" bson:"description"`
	HashedToken string     `json:"-" bson:"hashedToken"`
	Locked      *time.Time `json:"locked,omitempty" bson:"locked"`
}

func NewServiceAccount(id, description string) ServiceAccount {
	return ServiceAccount{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "ServiceAccount",
		},
		ObjectMeta: ObjectMeta{
			ID: id,
		},
		Description: description,
	}
}
