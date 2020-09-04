package queue

import "context"

type Writer interface {
	Write(ctx context.Context, message string) error
	Close(context.Context) error
}
