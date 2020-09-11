package queue

import "context"

type Reader interface {
	Read(ctx context.Context) (*Message, error)
	Close(context.Context) error
}
