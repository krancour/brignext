package sdk

import (
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
	"go.mongodb.org/mongo-driver/bson"
)

type ServiceAccount struct {
	meta.TypeMeta   `json:",inline" bson:",inline"`
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	Description     string     `json:"description" bson:"description"`
	HashedToken     string     `json:"-" bson:"hashedToken"`
	Locked          *time.Time `json:"locked,omitempty" bson:"locked"`
}

type ServiceAccountReference struct {
	meta.TypeMeta            `json:",inline"`
	meta.ObjectReferenceMeta `json:"metadata" bson:",inline"`
	Description              string     `json:"description" bson:"description"`
	Locked                   *time.Time `json:"locked" bson:"locked"`
}

func (s *ServiceAccountReference) UnmarshalBSON(bytes []byte) error {
	type ServiceAccountReferenceAlias ServiceAccountReference
	if err := bson.Unmarshal(
		bytes,
		&struct {
			*ServiceAccountReferenceAlias `bson:",inline"`
		}{
			ServiceAccountReferenceAlias: (*ServiceAccountReferenceAlias)(s),
		},
	); err != nil {
		return err
	}
	s.TypeMeta = meta.TypeMeta{
		APIVersion: meta.APIVersion,
		Kind:       "ServiceAccountReference",
	}
	return nil
}

type ServiceAccountReferenceList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`
	Items         []ServiceAccountReference `json:"items"`
}

func NewServiceAccountReferenceList() ServiceAccountReferenceList {
	return ServiceAccountReferenceList{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "ServiceAccountReferenceList",
		},
		ListMeta: meta.ListMeta{},
		Items:    []ServiceAccountReference{},
	}
}
