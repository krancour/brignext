package api

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type ServiceAccountReference struct {
	meta.ObjectReferenceMeta `json:"metadata"`
	Description              string     `json:"description"`
	Locked                   *time.Time `json:"locked,omitempty"`
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
