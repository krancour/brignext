package brignext

import (
	"time"
)

// nolint: lll
type ServiceAccount struct {
	TypeMeta           `json:",inline" bson:",inline"`
	ServiceAccountMeta `json:"metadata" bson:"metadata"`
	HashedToken        string                `json:"-" bson:"hashedToken"`
	Status             *ServiceAccountStatus `json:"status,omitempty" bson:"status"`
}

type ServiceAccountMeta struct {
	ID          string     `json:"id" bson:"id"`
	Description string     `json:"description" bson:"description"`
	Created     *time.Time `json:"created,omitempty" bson:"created"`
	// TODO: These fields are not yet in use
	CreatedBy     string     `json:"createdBy,omitempty" bson:"createdBy"`
	LastUpdated   *time.Time `json:"lastUpdated,omitempty" bson:"lastUpdated"`
	LastUpdatedBy string     `json:"lastUpdatedBy,omitempty" bson:"lastUpdatedBy"`
}

type ServiceAccountStatus struct {
	Locked bool `json:"locked" bson:"locked"`
}
