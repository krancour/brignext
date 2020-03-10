package redis

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

// backoff is used to delay a processing loop that has encountered an error.
// It calculates a backoff and sleeps for that amount of time.
func (c *consumer) backoff(
	ctx context.Context,
	err error,
	process string,
	failureCount uint8,
	maxFailures uint8,
	maxDelay time.Duration,
) {
	if failureCount >= maxFailures {
		c.abort(
			ctx,
			errors.Wrapf(err, "failed %d attempt(s) to %s", failureCount, process),
		)
	}
	delay := expBackoff(failureCount, maxDelay)
	timer := time.NewTimer(delay)
	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}

func (c *consumer) abort(ctx context.Context, err error) {
	select {
	case c.errCh <- err:
	case <-ctx.Done():
	}
}

// expBackoff implements a simple exponential backoff function.
func expBackoff(failureCount uint8, max time.Duration) time.Duration {
	base := math.Pow(2, float64(failureCount))
	jittered := (1 + rand.Float64()) * (base / 2)
	scaled := jittered * float64(time.Second)
	capped := math.Min(scaled, float64(max))
	return time.Duration(capped)
}
