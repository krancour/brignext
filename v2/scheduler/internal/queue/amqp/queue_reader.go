package amqp

import (
	"context"

	"github.com/Azure/go-amqp"
	"github.com/brigadecore/brigade/v2/scheduler/internal/queue"
	"github.com/pkg/errors"
)

type queueReader struct {
	queueName    string
	amqpSession  *amqp.Session
	amqpReceiver *amqp.Receiver
}

func (q *queueReader) Read(
	ctx context.Context,
) (*queue.Message, error) {
	amqpMsg, err := q.amqpReceiver.Receive(ctx)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error receiving AMQP message for queue %q",
			q.queueName,
		)
	}
	return &queue.Message{
		Message: string(amqpMsg.GetData()),
		Ack:     amqpMsg.Accept,
	}, nil
}

func (q *queueReader) Close(ctx context.Context) error {
	if err := q.amqpReceiver.Close(ctx); err != nil {
		return errors.Wrapf(
			err,
			"error closing AMQP receiver for queue %q",
			q.queueName,
		)
	}
	if err := q.amqpSession.Close(ctx); err != nil {
		return errors.Wrapf(
			err,
			"error closing AMQP session for queue %q",
			q.queueName,
		)
	}
	return nil
}
