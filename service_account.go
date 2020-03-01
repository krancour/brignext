package brignext

import (
	"time"
)

type ServiceAccount struct {
	ID          string     `json:"id,omitempty" bson:"_id"`
	Description string     `json:"description,omitempty" bson:"description"` //nolint: lll
	HashedToken string     `json:"-" bson:"hashedToken"`
	Created     *time.Time `json:"created,omitempty" bson:"created"`
	Locked      *bool      `json:"locked,omitempty" bson:"locked"`
}
