package redis

import (
	"github.com/go-redis/redis"
	"github.com/krancour/brignext/v2/pkg/messaging"
	"github.com/pkg/errors"
)

// producer is a Redis-based implementation of the messaging.Producer interface.
type producer struct {
	redisClient *redis.Client
	options     ProducerOptions
	// pendingListKey is the key for the global list of IDs for messages ready to
	// be handled.
	pendingListKey string
	// messagesHashKey is the key for the global hash of messages indexed by
	// message ID.
	messagesHashKey string
	// scheduledSetKey is the key for the global sorted set of IDs for messages
	// to be handled at or after some message-specific time in the future.
	scheduledSetKey string
}

// NewProducer returns a new Redis-based implementation of the
// messaging.Producer interface.
func NewProducer(
	baseQueueName string,
	redisClient *redis.Client,
	options *ProducerOptions,
) messaging.Producer {
	if options == nil {
		options = &ProducerOptions{}
	}
	return &producer{
		redisClient:     redisClient,
		options:         *options,
		pendingListKey:  pendingListKey(options.RedisPrefix, baseQueueName),
		messagesHashKey: messagesHashKey(options.RedisPrefix, baseQueueName),
		scheduledSetKey: scheduledSetKey(options.RedisPrefix, baseQueueName),
	}
}

func (p *producer) Publish(message messaging.Message) error {
	messageJSON, err := message.ToJSON()
	if err != nil {
		return errors.Wrapf(err, "error encoding message %q", message.ID())
	}

	pipeline := p.redisClient.TxPipeline()
	pipeline.HSet(p.messagesHashKey, message.ID(), messageJSON)

	if message.HandleTime() == nil {
		pipeline.LPush(p.pendingListKey, message.ID())
	} else {
		pipeline.ZAdd(
			p.scheduledSetKey,
			redis.Z{
				Score:  float64(message.HandleTime().Unix()),
				Member: message.ID(),
			},
		)
	}

	if _, err := pipeline.Exec(); err != nil {
		return errors.Wrapf(
			err,
			"error publishing message %q",
			message.ID(),
		)
	}

	return nil
}
