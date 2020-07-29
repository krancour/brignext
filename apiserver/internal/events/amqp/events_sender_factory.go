package amqp

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/Azure/go-amqp"
	"github.com/krancour/brignext/v2/apiserver/internal/events"
	"github.com/krancour/brignext/v2/internal/retries"
	"github.com/pkg/errors"
)

type eventsSenderFactory struct {
	address  string
	dialOpts []amqp.ConnOption
	// TODO: Replace this with some kind of options doohickey
	isAzureServiceBus bool
	amqpClient        *amqp.Client
	amqpClientMu      *sync.Mutex
}

func NewEventsSenderFactory(
	address string,
	username string,
	password string,
	isAzureServiceBus bool,
) (events.SenderFactory, error) {
	e := &eventsSenderFactory{
		address: address,
		dialOpts: []amqp.ConnOption{
			amqp.ConnSASLPlain(username, password),
		},
		isAzureServiceBus: isAzureServiceBus,
		amqpClientMu:      &sync.Mutex{},
	}
	if err := e.connect(); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *eventsSenderFactory) connect() error {
	return retries.ManageRetries(
		context.Background(),
		"connect",
		10,
		10*time.Second,
		func() (bool, error) {
			if e.amqpClient != nil {
				e.amqpClient.Close()
			}
			var err error
			if e.amqpClient, err = amqp.Dial(e.address, e.dialOpts...); err != nil {
				return true, errors.Wrap(err, "error dialing endpoint")
			}
			return false, nil
		},
	)
}

func (e *eventsSenderFactory) NewSender(
	projectID string,
) (events.Sender, error) {
	e.amqpClientMu.Lock()
	defer e.amqpClientMu.Unlock()

	linkOpts := []amqp.LinkOption{
		amqp.LinkTargetAddress(e.getLinkAddress(projectID)),
	}

	var amqpSession *amqp.Session
	var amqpSender *amqp.Sender
	var err error
	for {
		if amqpSession, err = e.amqpClient.NewSession(); err != nil {
			if err = e.connect(); err != nil {
				return nil, err
			}
			continue
		}
		if amqpSender, err = amqpSession.NewSender(linkOpts...); err != nil {
			amqpSession.Close(context.TODO())
			if err = e.connect(); err != nil {
				return nil, err
			}
			continue
		}
		break
	}

	return &eventsSender{
		isAzureServiceBus: e.isAzureServiceBus,
		projectID:         projectID,
		amqpSession:       amqpSession,
		amqpSender:        amqpSender,
	}, nil
}

func (e *eventsSenderFactory) Close(context.Context) error {
	if err := e.amqpClient.Close(); err != nil {
		return errors.Wrapf(err, "error closing AMQP client")
	}
	return nil
}

func (e *eventsSenderFactory) getLinkAddress(projectID string) string {
	if e.isAzureServiceBus {
		return "events" // TODO: Make base queue name configurable
	}
	return fmt.Sprintf("events.%s", projectID)
}
