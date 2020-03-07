package redis

import (
	"context"

	"github.com/go-redis/redis"
	"github.com/krancour/brignext/pkg/messaging"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// ConsumerOptions represents configutation options for a Consumer.
type ConsumerOptions struct {
	RedisPrefix  string
	WorkerCount  int
	WatcherCount int
}

// consumer is a Redis-based implementation of the messaging.Consumer interface.
type consumer struct {
	id            string
	baseQueueName string
	redisClient   *redis.Client
	options       ConsumerOptions
	// pendingQueueName is the global queue of messages ready to be handled.
	pendingQueueName string
	// deferredQueueName is the global queue of messages to be handled at or after
	// some message-specific time in the future.
	deferredQueueName string
	// consumersSetName is the global set of all consumers.
	consumersSetName string
	// activeQueueName is the set of messages actively being handled by this
	// consumer.
	activeQueueName string
	// watchedQueueName is the set of messages to be handled at or after some
	// message-specific time in the future, which have been placed into a holding
	// pattern by this consumer, which will add them to the pendingQueue after the
	// specified time has elapsed.
	watchedQueueName string
	// heartbeatKey is the key where this consumer will publish heartbeats and
	// where other consumers will look for proof of life.
	heartbeatKey string

	// All of the following behaviors can be overridden for testing purposes
	clean      func(context.Context) error
	cleanQueue func(
		ctx context.Context,
		consumerID string,
		sourceQueueName string,
		destinationQueueName string,
	) error
	runHeart               func(ctx context.Context) error
	heartbeat              func() error
	receivePendingMessages func(
		ctx context.Context,
		sourceQueueName string,
		destinationQueueName string,
		messageCh chan<- []byte,
		errCh chan<- error,
	)
	receiveDeferredMessages func(
		ctx context.Context,
		sourceQueueName string,
		destinationQueueName string,
		messageCh chan<- []byte,
		errCh chan<- error,
	)
	handleMessages func(
		ctx context.Context,
		messageCh <-chan []byte,
		handler messaging.HandlerFn,
		errCh chan<- error,
	)
	watchDeferredMessages func(
		ctx context.Context,
		messageCh <-chan []byte,
		errCh chan<- error,
	)
}

// NewConsumer returns a new Redis-based implementation of the
// messaging.Consumer interface.
func NewConsumer(
	baseQueueName string,
	redisClient *redis.Client,
	options *ConsumerOptions,
) messaging.Consumer {
	if options == nil {
		options = &ConsumerOptions{
			WorkerCount:  1,
			WatcherCount: 100,
		}
	}
	consumerID := uuid.NewV4().String()
	c := &consumer{
		id:                consumerID,
		baseQueueName:     baseQueueName,
		redisClient:       redisClient,
		options:           *options,
		pendingQueueName:  pendingQueueName(options.RedisPrefix, baseQueueName),
		deferredQueueName: deferredQueueName(options.RedisPrefix, baseQueueName),
		consumersSetName:  consumersSetName(options.RedisPrefix, baseQueueName),
		activeQueueName: activeQueueName(
			options.RedisPrefix,
			baseQueueName,
			consumerID,
		),
		watchedQueueName: watchedQueueName(
			options.RedisPrefix,
			baseQueueName,
			consumerID,
		),
		heartbeatKey: heartbeatKey(options.RedisPrefix, baseQueueName, consumerID),
	}
	c.clean = c.defaultClean
	c.cleanQueue = c.defaultCleanQueue
	c.runHeart = c.defaultRunHeart
	c.heartbeat = c.defaultHeartbeat
	c.receivePendingMessages = c.defaultReceiveMessages
	c.receiveDeferredMessages = c.defaultReceiveMessages
	c.handleMessages = c.defaultHandleMessages
	c.watchDeferredMessages = c.defaultWatchDeferredMessages
	return c
}

func (c *consumer) Consume(
	ctx context.Context,
	handler messaging.HandlerFn,
) error {
	// This function starts many goroutines. It is designed to exit if the context
	// it was passed is canceled or if any one of its constituent goroutines
	// completes. When it exits, context will be canceled and any remaining
	// goroutines should therefore receive their signal to shut down.

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// All goroutines we launch will send errors here
	errCh := make(chan error)

	// Start the cleaner
	go func() {
		err := c.clean(ctx)
		select {
		case errCh <- errors.Wrapf(
			err,
			"queue %q consumer %q cleaner stopped",
			c.baseQueueName,
			c.id,
		):
		case <-ctx.Done():
		}
	}()

	// As soon as we add the consumer to the consumers set, it's eligible for the
	// other consumers' cleaners to clean up after it, so it's important that we
	// guarantee that they will see this consumer as alive. We can't trust that
	// the heartbeat loop (which we'll shortly start in its own goroutine) will
	// have sent the first heartbeat BEFORE the consumer is added to the consumers
	// set. To head off that race condition, we synchronously send the first
	// heartbeat.
	if err := c.heartbeat(); err != nil {
		return errors.Wrapf(
			err,
			"error sending initial heartbeat for queue %q consumer %q",
			c.baseQueueName,
			c.id,
		)
	}

	// Heartbeat loop
	go func() {
		err := c.runHeart(ctx)
		select {
		case errCh <- errors.Wrapf(
			err,
			"queue %q consumer %q heart stopped",
			c.baseQueueName,
			c.id,
		):
		case <-ctx.Done():
		}
	}()

	// Announce this consumer's existence
	if err := c.redisClient.SAdd(c.consumersSetName, c.id).Err(); err != nil {
		return errors.Wrapf(
			err,
			"error adding consumer %q to queue %q consumers set",
			c.id,
			c.baseQueueName,
		)
	}

	// Assemble and execute a pipeline to receive and handle pending messages...
	go func() {
		messageCh := make(chan []byte)
		receiverErrCh := make(chan error)
		handlerErrCh := make(chan error)
		go c.receivePendingMessages(
			ctx,
			c.pendingQueueName,
			c.activeQueueName,
			messageCh,
			receiverErrCh,
		)
		// Fan out to desired number of message handlers
		for i := 0; i < c.options.WorkerCount; i++ {
			go c.handleMessages(
				ctx,
				messageCh,
				handler,
				handlerErrCh,
			)
		}
		select {
		case err := <-receiverErrCh:
			select {
			case errCh <- errors.Wrapf(
				err,
				"queue %q consumer %q pending message receiver stopped",
				c.baseQueueName,
				c.id,
			):
			case <-ctx.Done():
			}
		case err := <-handlerErrCh:
			select {
			case errCh <- errors.Wrapf(
				err,
				"queue %q consumer %q  message handler stopped",
				c.baseQueueName,
				c.id,
			):
			case <-ctx.Done():
			}
		case <-ctx.Done():
		}
	}()

	// Assemble and execute a pipeline to receive and watch deferred messages...
	go func() {
		messageCh := make(chan []byte)
		receiverErrCh := make(chan error)
		watcherErrCh := make(chan error)
		go c.receiveDeferredMessages(
			ctx,
			c.deferredQueueName,
			c.watchedQueueName,
			messageCh,
			receiverErrCh,
		)
		// Fan out to desired number of watchers
		for i := 0; i < c.options.WatcherCount; i++ {
			go c.watchDeferredMessages(
				ctx,
				messageCh,
				watcherErrCh,
			)
		}
		select {
		case err := <-receiverErrCh:
			select {
			case errCh <- errors.Wrapf(
				err,
				"queue %q consumer %q deferred message receiver stopped",
				c.baseQueueName,
				c.id,
			):
			case <-ctx.Done():
			}
		case err := <-watcherErrCh:
			select {
			case errCh <- errors.Wrapf(
				err,
				"queue %q consumer %q deferred message watcher stopped",
				c.baseQueueName,
				c.id,
			):
			case <-ctx.Done():
			}
		case <-ctx.Done():
		}
	}()

	// Now wait...
	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-errCh:
	}
	return errors.Wrapf(
		err,
		"queue %q consumer %q shutting down",
		c.baseQueueName,
		c.id,
	)
}
