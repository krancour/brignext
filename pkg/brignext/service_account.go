package brignext

import (
	"time"
)

type ServiceAccount struct {
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	Token       string    `json:"-" bson:"-"`
	Created     time.Time `json:"created" bson:"created"`
}
