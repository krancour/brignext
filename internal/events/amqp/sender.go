package amqp

import (
	"context"
	"log"

	amqp "github.com/Azure/go-amqp"
	"github.com/pkg/errors"
)

type sender struct {
	// TODO: Replace this with some kind of options doohickey
	isAzureServiceBus bool
	projectID         string
	amqpSession       *amqp.Session
	amqpSender        *amqp.Sender
}

func (s *sender) Send(ctx context.Context, event string) error {
	msg := &amqp.Message{
		Header: &amqp.MessageHeader{
			Durable: true,
		},
		Data: [][]byte{
			[]byte(event),
		},
	}
	if s.isAzureServiceBus {
		msg.Properties = &amqp.MessageProperties{
			GroupID: s.projectID,
		}
	}
	if err := s.amqpSender.Send(ctx, msg); err != nil {
		return errors.Wrapf(
			err,
			"error sending amqp message for project %q",
			s.projectID,
		)
	}
	return nil
}

func (s *sender) Close(ctx context.Context) error {
	if err := s.amqpSender.Close(ctx); err != nil {
		return errors.Wrapf(
			err,
			"error closing AMQP sender for project %q",
			s.projectID,
		)
	}
	if err := s.amqpSession.Close(ctx); err != nil {
		return errors.Wrapf(
			err,
			"error closing AMQP session for project %q",
			s.projectID,
		)
	}
	log.Printf(
		"DEBUG: closed AMQP-based event sender for project %q",
		s.projectID,
	)
	return nil
}
