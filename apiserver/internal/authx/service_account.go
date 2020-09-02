package authx

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/meta"
)

type ServiceAccount struct {
	meta.ObjectMeta     `json:"metadata" bson:",inline"`
	Description         string     `json:"description" bson:"description"`
	HashedToken         string     `json:"-" bson:"hashedToken"`
	Locked              *time.Time `json:"locked,omitempty" bson:"locked"`
	ServiceAccountRoles []Role     `json:"roles,omitempty" bson:"roles,omitempty"`
}

func (s *ServiceAccount) Roles() []Role {
	return s.ServiceAccountRoles
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

// ServiceAccountsSelector represents useful filter criteria when selecting
// multiple ServiceAccounts for API group operations like list. It currently has
// no fields, but exists for future expansion.
type ServiceAccountsSelector struct{}

// ServiceAccountList is an ordered and pageable list of ServiceAccounts.
type ServiceAccountList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of ServiceAccounts.
	Items []ServiceAccount `json:"items,omitempty"`
}

func (s ServiceAccountList) MarshalJSON() ([]byte, error) {
	type Alias ServiceAccountList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ServiceAccountList",
			},
			Alias: (Alias)(s),
		},
	)
}