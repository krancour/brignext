package brignext

import (
	"time"
)

type ServiceAccount struct {
	ID          string     `json:"id,omitempty" bson:"_id,omitempty"`
	Description string     `json:"description,omitempty" bson:"description,omitempty"` //nolint: lll
	HashedToken string     `json:"-" bson:"hashedToken,omitempty"`
	Created     *time.Time `json:"created,omitempty" bson:"created,omitempty"`
	Locked      *bool      `json:"locked,omitempty" bson:"locked,omitempty"`
}
