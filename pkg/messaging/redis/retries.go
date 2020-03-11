package redis

import (
	"context"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

func (c *consumer) manageRetries(
	ctx context.Context,
	process string,
	maxAttempts uint8,
	maxDelay time.Duration,
	fn func() error,
) bool {
	var attempts uint8
	var err error
	for attempts = 1; attempts <= maxAttempts; attempts++ {
		if err = fn(); err == nil {
			return true
		}
		log.Printf(
			"WARNING: queue %q consumer %q failed to %s; will retry: %s",
			c.baseQueueName,
			c.id,
			process,
			err,
		)
		select {
		case <-time.After(expBackoff(attempts, maxDelay)):
		case <-ctx.Done():
		}
	}
	err = errors.Wrapf(
		err,
		"queue %q consumer %q failed %d attempt(s) to %s",
		c.baseQueueName,
		c.id,
		maxAttempts,
		process,
	)
	select {
	case c.errCh <- err:
	case <-ctx.Done():
	}
	return false
}

// expBackoff implements a simple exponential backoff function.
func expBackoff(failureCount uint8, max time.Duration) time.Duration {
	base := math.Pow(2, float64(failureCount))
	// TODO: This rand isn't seeded correctly
	jittered := (1 + rand.Float64()) * (base / 2)
	scaled := jittered * float64(time.Second)
	capped := math.Min(scaled, float64(max))
	return time.Duration(capped)
}
