package redis

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/krancour/brignext/pkg/rand"
	"github.com/pkg/errors"
)

var seededRand = rand.NewSeeded()

func (c *consumer) manageRetries(
	ctx context.Context,
	process string,
	fn func() error,
) bool {
	var failedAttempts uint8
	var err error
	for {
		if err = fn(); err == nil {
			return true
		}
		failedAttempts++
		if failedAttempts == *c.options.RedisOperationMaxAttempts {
			break
		}
		delay := jitteredExpBackoff(failedAttempts, *c.options.RedisOperationMaxBackoff)
		log.Printf(
			"WARNING: queue %q consumer %q failed %d attempts(s) to %s; will "+
				"retry in %s: %s",
			c.queueName,
			c.id,
			failedAttempts,
			process,
			delay,
			err,
		)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
		}
	}
	err = errors.Wrapf(
		err,
		"queue %q consumer %q failed %d attempt(s) to %s",
		c.queueName,
		c.id,
		*c.options.RedisOperationMaxAttempts,
		process,
	)
	select {
	case c.errCh <- err:
	case <-ctx.Done():
	}
	return false
}

func jitteredExpBackoff(
	failureCount uint8,
	maxDelay time.Duration,
) time.Duration {
	base := math.Pow(2, float64(failureCount))
	capped := math.Min(base, maxDelay.Seconds())
	jittered := (1 + seededRand.Float64()) * (capped / 2)
	scaled := jittered * float64(time.Second)
	return time.Duration(scaled)
}

func expBackoff(failureCount uint8, maxDelay time.Duration) time.Duration {
	base := math.Pow(2, float64(failureCount))
	capped := math.Min(base, maxDelay.Seconds())
	scaled := capped * float64(time.Second)
	return time.Duration(scaled)
}

func maxCummulativeBackoff(
	maxTries uint8,
	maxDelay time.Duration,
) time.Duration {
	var sum time.Duration
	var i uint8
	for i = 1; i < maxTries; i++ {
		sum += expBackoff(i, maxDelay)
	}
	return sum
}
