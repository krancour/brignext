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
	var attempts uint8
	var err error
	for attempts = 1; attempts <= *c.options.RedisOperationMaxAttempts; attempts++ { // nolint: lll
		if err = fn(); err == nil {
			return true
		}
		log.Printf(
			"WARNING: queue %q consumer %q failed to %s; will retry: %s",
			c.queueName,
			c.id,
			process,
			err,
		)
		select {
		case <-time.After(expBackoff(attempts, *c.options.RedisOperationMaxBackoff)): // nolint: lll
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

func expBackoff(failureCount uint8, max time.Duration) time.Duration {
	base := math.Pow(2, float64(failureCount))
	jittered := (1 + seededRand.Float64()) * (base / 2)
	scaled := jittered * float64(time.Second)
	capped := math.Min(scaled, float64(max))
	return time.Duration(capped)
}
