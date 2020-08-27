package amqp

import (
	"context"

	amqp "github.com/Azure/go-amqp"
	"github.com/pkg/errors"
)

type queueWriter struct {
	queueName   string
	groupID     string
	amqpSession *amqp.Session
	amqpSender  *amqp.Sender
}

func (q *queueWriter) Write(ctx context.Context, message string) error {
	msg := &amqp.Message{
		Header: &amqp.MessageHeader{
			Durable: true,
		},
		Data: [][]byte{
			[]byte(message),
		},
	}
	if q.groupID != "" {
		msg.Properties = &amqp.MessageProperties{
			GroupID: q.groupID,
		}
	}
	if err := q.amqpSender.Send(ctx, msg); err != nil {
		return errors.Wrapf(
			err,
			"error sending amqp message for queue %q",
			q.queueName,
		)
	}
	return nil
}

func (q *queueWriter) Close(ctx context.Context) error {
	if err := q.amqpSender.Close(ctx); err != nil {
		return errors.Wrapf(
			err,
			"error closing AMQP sender for queue %q",
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
