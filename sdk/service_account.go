package sdk

import (
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type ServiceAccountList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`
	Items         []ServiceAccount `json:"items"`
}

func NewServiceAccountList() ServiceAccountList {
	return ServiceAccountList{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "ServiceAccountList",
		},
		Items: []ServiceAccount{},
	}
}

type ServiceAccount struct {
	meta.TypeMeta   `json:",inline" bson:",inline"`
	meta.ObjectMeta `json:"metadata" bson:"metadata"`
	Description     string     `json:"description" bson:"description"`
	HashedToken     string     `json:"-" bson:"hashedToken"`
	Locked          *time.Time `json:"locked,omitempty" bson:"locked"`
}

func NewServiceAccount(id, description string) ServiceAccount {
	return ServiceAccount{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "ServiceAccount",
		},
		ObjectMeta: meta.ObjectMeta{
			ID: id,
		},
		Description: description,
	}
}
