package brignext

import (
	"time"
)

type ServiceAccount struct {
	ID          string    `json:"-" bson:"_id,omitempty"`
	Name        string    `json:"name,omitempty" bson:"name,omitempty"`
	Description string    `json:"description,omitempty" bson:"description,omitempty"`
	Created     time.Time `json:"created,omitempty" bson:"created,omitempty"`
}
