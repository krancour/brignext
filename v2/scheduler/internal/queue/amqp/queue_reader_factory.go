package amqp

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	amqp "github.com/Azure/go-amqp"
	"github.com/brigadecore/brigade/v2/internal/retries"
	"github.com/brigadecore/brigade/v2/scheduler/internal/queue"
	"github.com/pkg/errors"
)

type queueReaderFactory struct {
	address  string
	dialOpts []amqp.ConnOption
	// TODO: Replace this field with some kind of options struct to allow for
	// future expansion
	isAzureServiceBus bool
	amqpClient        *amqp.Client
	amqpClientMu      *sync.Mutex
}

func NewQueueReaderFactory(
	address string,
	username string,
	password string,
	isAzureServiceBus bool,
) (queue.ReaderFactory, error) {
	q := &queueReaderFactory{
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

func (q *queueReaderFactory) connect() error {
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

func (q *queueReaderFactory) NewQueueReader(
	queueName string,
) (queue.Reader, error) {
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
		amqp.LinkSourceAddress(queueName),
		// Link credit is 1 because we're a "slow" consumer. We do not want messages
		// piling up in a client-side buffer, knowing that it could be some time
		// before we can process them.
		amqp.LinkCredit(1),
	}
	if q.isAzureServiceBus {
		linkOpts = append(linkOpts, q.linkAzureSesionFilter(groupID))
	}

	var amqpSession *amqp.Session
	var amqpReceiver *amqp.Receiver
	var err error
	for {
		if amqpSession, err = q.amqpClient.NewSession(); err != nil {
			if err = q.connect(); err != nil {
				return nil, err
			}
			continue
		}
		if amqpReceiver, err = amqpSession.NewReceiver(linkOpts...); err != nil {
			amqpSession.Close(context.TODO())
			if err = q.connect(); err != nil {
				return nil, err
			}
			continue
		}
		break
	}

	return &queueReader{
		queueName:    queueName,
		amqpSession:  amqpSession,
		amqpReceiver: amqpReceiver,
	}, nil
}

func (q *queueReaderFactory) Close(context.Context) error {
	if err := q.amqpClient.Close(); err != nil {
		return errors.Wrapf(err, "error closing AMQP client")
	}
	log.Println("DEBUG: closed AMQP-based queue reader factory")
	return nil
}

func (q *queueReaderFactory) linkAzureSesionFilter(
	sessionID string,
) amqp.LinkOption {
	const name = "com.microsoft:session-filter"
	const code = uint64(0x00000137000000C)
	return amqp.LinkSourceFilter(name, code, sessionID)
}
