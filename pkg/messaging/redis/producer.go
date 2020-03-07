package redis

import (
	"github.com/go-redis/redis"
	"github.com/krancour/brignext/pkg/messaging"
	"github.com/pkg/errors"
)

// ProducerOptions represents configutation options for a Producer.
type ProducerOptions struct {
	RedisPrefix string
}

// producer is a Redis-based implementation of the messaging.Producer interface.
type producer struct {
	redisClient       *redis.Client
	options           ProducerOptions
	pendingQueueName  string
	deferredQueueName string
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
		redisClient:       redisClient,
		options:           *options,
		pendingQueueName:  pendingQueueName(options.RedisPrefix, baseQueueName),
		deferredQueueName: deferredQueueName(options.RedisPrefix, baseQueueName),
	}
}

func (p *producer) Publish(message messaging.Message) error {
	messageJSON, err := message.ToJSON()
	if err != nil {
		return errors.Wrapf(err, "error encoding message %q", message.ID())
	}

	var queueName string
	if message.HandleTime() != nil {
		queueName = p.deferredQueueName
	} else {
		queueName = p.pendingQueueName
	}

	if err = p.redisClient.LPush(queueName, messageJSON).Err(); err != nil {
		return errors.Wrapf(
			err,
			"error submitting message %q to queue %q",
			message.ID(),
			queueName,
		)
	}
	return nil
}
