package brignext

import (
	"time"
)

type ServiceAccount struct {
	ID          string     `json:"id" bson:"_id"`
	Description string     `json:"description" bson:"description"`
	HashedToken string     `json:"-" bson:"hashedToken"`
	Created     *time.Time `json:"created,omitempty" bson:"created"`
	Locked      *bool      `json:"locked,omitempty" bson:"locked"`
}
