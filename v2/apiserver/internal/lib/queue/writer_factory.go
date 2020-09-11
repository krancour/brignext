package queue

import "context"

type WriterFactory interface {
	NewQueueWriter(queueName string) (Writer, error)
	Close(context.Context) error
}
