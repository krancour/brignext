package brignext

import (
	"time"
)

type ServiceAccount struct {
	Name        string    `json:"name,omitempty" bson:"name,omitempty"`
	Description string    `json:"description,omitempty" bson:"description,omitempty"`
	Token       string    `json:"-" bson:"-"`
	Created     time.Time `json:"created,omitempty" bson:"created,omitempty"`
}
