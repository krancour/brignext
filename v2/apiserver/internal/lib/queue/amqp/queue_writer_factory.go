package amqp

import (
	"context"
	"strings"
	"sync"
	"time"

	amqp "github.com/Azure/go-amqp"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue"
	"github.com/brigadecore/brigade/v2/internal/retries"
	"github.com/pkg/errors"
)

type queueWriterFactory struct {
	address  string
	dialOpts []amqp.ConnOption
	// TODO: Replace this with some kind of options doohickey
	isAzureServiceBus bool
	amqpClient        *amqp.Client
	amqpClientMu      *sync.Mutex
}

func NewQueueWriterFactory(
	address string,
	username string,
	password string,
	isAzureServiceBus bool,
) (queue.WriterFactory, error) {
	q := &queueWriterFactory{
		address: address,
		dialOpts: []amqp.ConnOption{
			amqp.ConnSASLPlain(username, password),
		},
		isAzureServiceBus: isAzureServiceBus,
		amqpClientMu:      &sync.Mutex{},
	}
	if err := q.connect(); err != nil {
		return nil, err
	}
	return q, nil
}

func (q *queueWriterFactory) connect() error {
	return retries.ManageRetries(
		context.Background(),
		"connect",
		10,
		10*time.Second,
		func() (bool, error) {
			if q.amqpClient != nil {
				q.amqpClient.Close()
			}
			var err error
			if q.amqpClient, err = amqp.Dial(q.address, q.dialOpts...); err != nil {
				return true, errors.Wrap(err, "error dialing endpoint")
			}
			return false, nil
		},
	)
}

func (q *queueWriterFactory) NewQueueWriter(
	queueName string,
) (queue.Writer, error) {
	q.amqpClientMu.Lock()
	defer q.amqpClientMu.Unlock()

	// Azure Service Bus allows only a relatively small number of queues per bus
	// so we want to conserve these precious resources. If the queue name has one
	// and only one "." character in it, we split the queue name into a "base
	// name" (everything to the left of the ".") and a "group ID" (everything to
	// the the right of the "."). Despite not allowing large numbers of queues per
	// bus, Azure Service Bus can multiplex queues based on a "group ID"
	// (sometimes aka "session ID," but we shun that terminology here because that
	// would overload the term "session," which has specific meaning in the
	// context of AMQP.) This multiplexing emulates roughly the same behavior as
	// using a queue per group ID.
	var groupID string
	if q.isAzureServiceBus {
		queueNameTokens := strings.Split(queueName, ".")
		if len(queueNameTokens) == 2 {
			queueName = queueNameTokens[0]
			groupID = queueNameTokens[1]
		}
	}

	linkOpts := []amqp.LinkOption{
		amqp.LinkTargetAddress(queueName),
	}

	var amqpSession *amqp.Session
	var amqpSender *amqp.Sender
	var err error
	for {
		if amqpSession, err = q.amqpClient.NewSession(); err != nil {
			if err = q.connect(); err != nil {
				return nil, err
			}
			continue
		}
		if amqpSender, err = amqpSession.NewSender(linkOpts...); err != nil {
			amqpSession.Close(context.TODO())
			if err = q.connect(); err != nil {
				return nil, err
			}
			continue
		}
		break
	}

	return &queueWriter{
		queueName:   queueName,
		groupID:     groupID,
		amqpSession: amqpSession,
		amqpSender:  amqpSender,
	}, nil
}

func (q *queueWriterFactory) Close(context.Context) error {
	if err := q.amqpClient.Close(); err != nil {
		return errors.Wrapf(err, "error closing AMQP client")
	}
	return nil
}
