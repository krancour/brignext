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

// ServiceAccountListOptions represents useful filter criteria when selecting
// multiple ServiceAccounts for API group operations like list.
type ServiceAccountListOptions struct {
	// Continue aids in pagination of long lists. It permits clients to echo an
	// opaque value obtained from a previous API call back to the API in a
	// subsequent call in order to indicate what resource was the last on the
	// previous page.
	Continue string
	// Limit aids in pagination of long lists. It permits clients to specify page
	// size when making API calls. The API server provides a default when a value
	// is not specified and may reject or override invalid values (non-positive)
	// numbers or very large page sizes.
	Limit int64
}

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
