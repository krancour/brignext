package events

import "context"

type Sender interface {
	Send(ctx context.Context, event string) error
	Close(context.Context) error
}
