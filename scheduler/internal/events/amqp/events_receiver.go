package amqp

import (
	"context"

	"github.com/Azure/go-amqp"
	"github.com/krancour/brignext/v2/scheduler/internal/events"
	"github.com/pkg/errors"
)

type eventsReceiver struct {
	projectID    string
	amqpSession  *amqp.Session
	amqpReceiver *amqp.Receiver
}

func (e *eventsReceiver) Receive(ctx context.Context) (*events.AsyncEvent, error) {
	amqpMsg, err := e.amqpReceiver.Receive(ctx)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error receiving AMQP message for project %q",
			e.projectID,
		)
	}
	return &events.AsyncEvent{
		EventID: string(amqpMsg.GetData()),
		Ack:     amqpMsg.Accept,
	}, nil
}

func (e *eventsReceiver) Close(ctx context.Context) error {
	if err := e.amqpReceiver.Close(ctx); err != nil {
		return errors.Wrapf(
			err,
			"error closing AMQP receiver for project %q",
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
