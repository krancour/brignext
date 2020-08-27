package queue

import "context"

type ReaderFactory interface {
	NewQueueReader(queueName string) (Reader, error)
	Close(context.Context) error
}
