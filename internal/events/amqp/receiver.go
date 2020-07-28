package amqp

import (
	"context"
	"log"

	"github.com/Azure/go-amqp"
	"github.com/krancour/brignext/v2/internal/events"
	"github.com/pkg/errors"
)

type receiver struct {
	projectID    string
	amqpSession  *amqp.Session
	amqpReceiver *amqp.Receiver
}

func (r *receiver) Receive(ctx context.Context) (*events.AsyncEvent, error) {
	amqpMsg, err := r.amqpReceiver.Receive(ctx)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error receiving AMQP message for project %q",
			r.projectID,
		)
	}
	return &events.AsyncEvent{
		EventID: string(amqpMsg.GetData()),
		Ack:     amqpMsg.Accept,
	}, nil
}

func (r *receiver) Close(ctx context.Context) error {
	if err := r.amqpReceiver.Close(ctx); err != nil {
		return errors.Wrapf(
			err,
			"error closing AMQP receiver for project %q",
			r.projectID,
		)
	}
	if err := r.amqpSession.Close(ctx); err != nil {
		return errors.Wrapf(
			err,
			"error closing AMQP session for project %q",
			r.projectID,
		)
	}
	log.Printf(
		"DEBUG: closed AMQP-based event receiver for project %q",
		r.projectID,
	)
	return nil
}
