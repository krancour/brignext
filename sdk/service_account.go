package sdk

import (
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type ServiceAccount struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata"`
	Description     string     `json:"description"`
	Locked          *time.Time `json:"locked,omitempty"`
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

type ServiceAccountReference struct {
	meta.TypeMeta            `json:",inline"`
	meta.ObjectReferenceMeta `json:"metadata"`
	Description              string     `json:"description"`
	Locked                   *time.Time `json:"locked,omitempty"`
}

type ServiceAccountReferenceList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`
	Items         []ServiceAccountReference `json:"items"`
}
