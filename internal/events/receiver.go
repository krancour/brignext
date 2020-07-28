package events

import "context"

type Receiver interface {
	Receive(ctx context.Context) (*AsyncEvent, error)
	Close(context.Context) error
}
