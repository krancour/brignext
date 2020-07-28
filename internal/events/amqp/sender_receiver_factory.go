package amqp

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/Azure/go-amqp"
	"github.com/krancour/brignext/v2/internal/events"
	"github.com/krancour/brignext/v2/internal/retries"
	"github.com/pkg/errors"
)

type senderReceiverFactory struct {
	address  string
	dialOpts []amqp.ConnOption
	// TODO: Replace this with some kind of options doohickey
	isAzureServiceBus bool
	amqpClient        *amqp.Client
	amqpClientMu      *sync.Mutex
}

func NewReceiverFactory(
	address string,
	username string,
	password string,
	isAzureServiceBus bool,
) (events.ReceiverFactory, error) {
	return newSenderReceiverFactory(
		address,
		username,
		password,
		isAzureServiceBus,
	)
}

func NewSenderFactory(
	address string,
	username string,
	password string,
	isAzureServiceBus bool,
) (events.SenderFactory, error) {
	return newSenderReceiverFactory(
		address,
		username,
		password,
		isAzureServiceBus,
	)
}

func newSenderReceiverFactory(
	address string,
	username string,
	password string,
	isAzureServiceBus bool,
) (*senderReceiverFactory, error) {
	s := &senderReceiverFactory{
		address: address,
		dialOpts: []amqp.ConnOption{
			amqp.ConnSASLPlain(username, password),
		},
		isAzureServiceBus: isAzureServiceBus,
		amqpClientMu:      &sync.Mutex{},
	}
	if err := s.connect(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *senderReceiverFactory) connect() error {
	return retries.ManageRetries(
		context.Background(),
		"connect",
		10,
		10*time.Second,
		func() (bool, error) {
			if s.amqpClient != nil {
				s.amqpClient.Close()
			}
			var err error
			if s.amqpClient, err = amqp.Dial(s.address, s.dialOpts...); err != nil {
				return true, errors.Wrap(err, "error dialing endpoint")
			}
			return false, nil
		},
	)
}

func (s *senderReceiverFactory) NewSender(
	projectID string,
) (events.Sender, error) {
	s.amqpClientMu.Lock()
	defer s.amqpClientMu.Unlock()

	linkOpts := []amqp.LinkOption{
		amqp.LinkTargetAddress(s.getLinkAddress(projectID)),
	}

	var amqpSession *amqp.Session
	var amqpSender *amqp.Sender
	var err error
	for {
		if amqpSession, err = s.amqpClient.NewSession(); err != nil {
			if err = s.connect(); err != nil {
				return nil, err
			}
			continue
		}
		if amqpSender, err = amqpSession.NewSender(linkOpts...); err != nil {
			amqpSession.Close(context.TODO())
			if err = s.connect(); err != nil {
				return nil, err
			}
			continue
		}
		break
	}

	log.Printf(
		"DEBUG: created AMQP-based event receiver for project %q",
		projectID,
	)

	return &sender{
		isAzureServiceBus: s.isAzureServiceBus,
		projectID:         projectID,
		amqpSession:       amqpSession,
		amqpSender:        amqpSender,
	}, nil
}

func (s *senderReceiverFactory) NewReceiver(
	projectID string,
) (events.Receiver, error) {
	s.amqpClientMu.Lock()
	defer s.amqpClientMu.Unlock()

	linkOpts := []amqp.LinkOption{
		amqp.LinkSourceAddress(s.getLinkAddress(projectID)),
		// Link credit is 1 because we're a "slow" consumer. We do not want messages
		// piling up in a client-side buffer, knowing that it could be some time
		// before we can process them.
		amqp.LinkCredit(1),
	}
	if s.isAzureServiceBus {
		linkOpts = append(linkOpts, s.linkAzureSesionFilter(projectID))
	}

	var amqpSession *amqp.Session
	var amqpReceiver *amqp.Receiver
	var err error
	for {
		if amqpSession, err = s.amqpClient.NewSession(); err != nil {
			if err = s.connect(); err != nil {
				return nil, err
			}
			continue
		}
		if amqpReceiver, err = amqpSession.NewReceiver(linkOpts...); err != nil {
			amqpSession.Close(context.TODO())
			if err = s.connect(); err != nil {
				return nil, err
			}
			continue
		}
		break
	}

	log.Printf(
		"DEBUG: created AMQP-based event receiver for project %q",
		projectID,
	)

	return &receiver{
		projectID:    projectID,
		amqpSession:  amqpSession,
		amqpReceiver: amqpReceiver,
	}, nil
}

func (s *senderReceiverFactory) Close(context.Context) error {
	if err := s.amqpClient.Close(); err != nil {
		return errors.Wrapf(err, "error closing AMQP client")
	}
	log.Println("DEBUG: closed AMQP-based event sender/receiver factory")
	return nil
}

func (s *senderReceiverFactory) getLinkAddress(projectID string) string {
	if s.isAzureServiceBus {
		return "events"
	}
	return fmt.Sprintf("events.%s", projectID)
}

func (s *senderReceiverFactory) linkAzureSesionFilter(
	sessionID string,
) amqp.LinkOption {
	const name = "com.microsoft:session-filter"
	const code = uint64(0x00000137000000C)
	return amqp.LinkSourceFilter(name, code, sessionID)
}
