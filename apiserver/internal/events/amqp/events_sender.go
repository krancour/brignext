package amqp

import (
	"context"

	amqp "github.com/Azure/go-amqp"
	"github.com/pkg/errors"
)

type eventsSender struct {
	// TODO: Replace this with some kind of options doohickey
	isAzureServiceBus bool
	projectID         string
	amqpSession       *amqp.Session
	amqpSender        *amqp.Sender
}

func (e *eventsSender) Send(ctx context.Context, event string) error {
	msg := &amqp.Message{
		Header: &amqp.MessageHeader{
			Durable: true,
		},
		Data: [][]byte{
			[]byte(event),
		},
	}
	if e.isAzureServiceBus {
		msg.Properties = &amqp.MessageProperties{
			GroupID: e.projectID,
		}
	}
	if err := e.amqpSender.Send(ctx, msg); err != nil {
		return errors.Wrapf(
			err,
			"error sending amqp message for project %q",
			e.projectID,
		)
	}
	return nil
}

func (e *eventsSender) Close(ctx context.Context) error {
	if err := e.amqpSender.Close(ctx); err != nil {
		return errors.Wrapf(
			err,
			"error closing AMQP sender for project %q",
			e.projectID,
		)
	}
	if err := e.amqpSession.Close(ctx); err != nil {
		return errors.Wrapf(
			err,
			"error closing AMQP session for project %q",
			e.projectID,
		)
	}
	return nil
}
