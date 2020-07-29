package amqp

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/Azure/go-amqp"
	"github.com/krancour/brignext/v2/controller/internal/events"
	"github.com/krancour/brignext/v2/internal/retries"
	"github.com/pkg/errors"
)

type eventsReceiverFactory struct {
	address  string
	dialOpts []amqp.ConnOption
	// TODO: Replace this with some kind of options doohickey
	isAzureServiceBus bool
	amqpClient        *amqp.Client
	amqpClientMu      *sync.Mutex
}

func NewEventsReceiverFactory(
	address string,
	username string,
	password string,
	isAzureServiceBus bool,
) (events.ReceiverFactory, error) {
	e := &eventsReceiverFactory{
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

func (e *eventsReceiverFactory) connect() error {
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

func (e *eventsReceiverFactory) NewReceiver(
	projectID string,
) (events.Receiver, error) {
	e.amqpClientMu.Lock()
	defer e.amqpClientMu.Unlock()

	linkOpts := []amqp.LinkOption{
		amqp.LinkSourceAddress(e.getLinkAddress(projectID)),
		// Link credit is 1 because we're a "slow" consumer. We do not want messages
		// piling up in a client-side buffer, knowing that it could be some time
		// before we can process them.
		amqp.LinkCredit(1),
	}
	if e.isAzureServiceBus {
		linkOpts = append(linkOpts, e.linkAzureSesionFilter(projectID))
	}

	var amqpSession *amqp.Session
	var amqpReceiver *amqp.Receiver
	var err error
	for {
		if amqpSession, err = e.amqpClient.NewSession(); err != nil {
			if err = e.connect(); err != nil {
				return nil, err
			}
			continue
		}
		if amqpReceiver, err = amqpSession.NewReceiver(linkOpts...); err != nil {
			amqpSession.Close(context.TODO())
			if err = e.connect(); err != nil {
				return nil, err
			}
			continue
		}
		break
	}

	return &eventsReceiver{
		projectID:    projectID,
		amqpSession:  amqpSession,
		amqpReceiver: amqpReceiver,
	}, nil
}

func (e *eventsReceiverFactory) Close(context.Context) error {
	if err := e.amqpClient.Close(); err != nil {
		return errors.Wrapf(err, "error closing AMQP client")
	}
	log.Println("DEBUG: closed AMQP-based event sender/receiver factory")
	return nil
}

func (e *eventsReceiverFactory) getLinkAddress(projectID string) string {
	if e.isAzureServiceBus {
		return "events"
	}
	return fmt.Sprintf("events.%s", projectID)
}

func (e *eventsReceiverFactory) linkAzureSesionFilter(
	sessionID string,
) amqp.LinkOption {
	const name = "com.microsoft:session-filter"
	const code = uint64(0x00000137000000C)
	return amqp.LinkSourceFilter(name, code, sessionID)
}
