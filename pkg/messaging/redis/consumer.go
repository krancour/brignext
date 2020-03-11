package redis

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/krancour/brignext/pkg/messaging"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// consumer is a Redis-based implementation of the messaging.Consumer interface.
type consumer struct {
	id          string
	redisClient *redis.Client
	queueName   string
	options     ConsumerOptions
	handler     messaging.HandlerFn

	// pendingListKey is the key for the global list of IDs for messages ready to
	// be handled.
	pendingListKey string
	// messagesHashKey is the key for the global hash of messages indexed by
	// message ID.
	messagesHashKey string
	// scheduledSetKey is the key for the global sorted set of IDs for messages
	// to be handled at or after some message-specific time in the future.
	scheduledSetKey string
	// consumersSetKey is the key for the global set of all consumers.
	consumersSetKey string
	// activeListKey is the key for the list of messages actively being handled
	// by this consumer.
	activeListKey string

	// Scripts
	schedulerScriptSHA string
	cleanerScriptSHA   string

	// All of the following behaviors can be overridden for testing purposes
	runHeart               func(context.Context)
	runCleaner             func(context.Context)
	receivePendingMessages func(context.Context)
	handleMessages         func(context.Context)
	runScheduler           func(context.Context)

	handlerReadyCh chan struct{}
	messageCh      chan messaging.Message
	// All goroutines we launch will send errors here
	errCh chan error

	wg *sync.WaitGroup
}

// NewConsumer returns a new Redis-based implementation of the
// messaging.Consumer interface.
func NewConsumer(
	redisClient *redis.Client,
	queueName string,
	options *ConsumerOptions,
	handler messaging.HandlerFn,
) (messaging.Consumer, error) {
	if options == nil {
		options = &ConsumerOptions{}
	}
	options.applyDefaults()
	consumerID := uuid.NewV4().String()
	c := &consumer{
		id:              consumerID,
		redisClient:     redisClient,
		queueName:       queueName,
		options:         *options,
		handler:         handler,
		pendingListKey:  pendingListKey(options.RedisPrefix, queueName),
		messagesHashKey: messagesHashKey(options.RedisPrefix, queueName),
		scheduledSetKey: scheduledSetKey(options.RedisPrefix, queueName),
		consumersSetKey: consumersSetKey(options.RedisPrefix, queueName),
		activeListKey: activeListKey(
			options.RedisPrefix,
			queueName,
			consumerID,
		),
		handlerReadyCh: make(chan struct{}),
		messageCh:      make(chan messaging.Message),
		errCh:          make(chan error),
		wg:             &sync.WaitGroup{},
	}

	var err error

	// Scheduler script
	c.schedulerScriptSHA, err = redisClient.ScriptLoad(schedulerScript).Result()
	if err != nil {
		return nil, errors.Wrap(err, "error loading scheduler script")
	}

	// Cleaner script
	c.cleanerScriptSHA, err = redisClient.ScriptLoad(cleanerScript).Result()
	if err != nil {
		return nil, errors.Wrap(err, "error loading cleaner script")
	}

	// Behaviors
	c.runCleaner = c.defaultRunCleaner
	c.runHeart = c.defaultRunHeart
	c.receivePendingMessages = c.defaultReceivePendingMessages
	c.handleMessages = c.defaultHandleMessages
	c.runScheduler = c.defaultRunScheduler

	return c, nil
}

func (c *consumer) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Send the first heartbeat synchronously before we doing anything else so
	// that over-eager cleaners belonging to other consumers of the same reliable
	// queue won't think us dead while we're still initializing.
	if err := c.heartbeat(); err != nil {
		return errors.Wrapf(
			err,
			"error sending initial heartbeat for queue %q consumer %q",
			c.queueName,
			c.id,
		)
	}

	c.wg.Add(3)
	// Start the heartbeat loop
	go c.runHeart(ctx)
	// Start the cleaner loop
	go c.runCleaner(ctx)
	// Move scheduled tasks to the pending list as they become ready
	go c.runScheduler(ctx)

	// Fan out to desired number of message receivers
	c.wg.Add(int(*c.options.ConcurrentReceiversCount))
	var i uint8
	for i = 0; i < *c.options.ConcurrentReceiversCount; i++ {
		go c.receivePendingMessages(ctx)
	}

	// Fan out to desired number of message handlers
	c.wg.Add(int(*c.options.ConcurrentHandlersCount))
	for i = 0; i < *c.options.ConcurrentHandlersCount; i++ {
		go c.handleMessages(ctx)
	}

	// Wait for an error or a completed context
	var err error
	select {
	case err = <-c.errCh:
		cancel() // Shut it all down
	case <-ctx.Done():
	}

	// Wait for everything to finish
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		c.wg.Wait()
	}()
	select {
	case <-doneCh:
	case <-time.After(*c.options.ShutdownGracePeriod):
	}

	return err
}
