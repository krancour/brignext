package sdk

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
)

type ServiceAccount struct {
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	Description     string     `json:"description" bson:"description"`
	HashedToken     string     `json:"-" bson:"hashedToken"`
	Locked          *time.Time `json:"locked,omitempty" bson:"locked"`
}

func (s ServiceAccount) MarshalJSON() ([]byte, error) {
	type Alias ServiceAccount
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ServiceAccount",
			},
			Alias: (Alias)(s),
		},
	)
}

type ServiceAccountReference struct {
	meta.ObjectReferenceMeta `json:"metadata" bson:",inline"`
	Description              string     `json:"description" bson:"description"`
	Locked                   *time.Time `json:"locked" bson:"locked"`
}

func (s ServiceAccountReference) MarshalJSON() ([]byte, error) {
	type Alias ServiceAccountReference
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ServiceAccountReference",
			},
			Alias: (Alias)(s),
		},
	)
}

type ServiceAccountReferenceList struct {
	Items []ServiceAccountReference `json:"items"`
}

func (s ServiceAccountReferenceList) MarshalJSON() ([]byte, error) {
	type Alias ServiceAccountReferenceList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ServiceAccountReferenceList",
			},
			Alias: (Alias)(s),
		},
	)
}
