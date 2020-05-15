package retries

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/krancour/brignext/v2/internal/common/rand"
	"github.com/pkg/errors"
)

var seededRand = rand.NewSeeded()

func ManageRetries(
	ctx context.Context,
	process string,
	maxAttempts uint8,
	maxBackoff time.Duration,
	fn func() (bool, error),
) error {
	var failedAttempts uint8
	for {
		retry, err := fn()
		if !retry {
			return err
		}
		failedAttempts++
		if failedAttempts == maxAttempts {
			return errors.Wrapf(
				err,
				"failed %d attempt(s) to %s",
				maxAttempts,
				process,
			)
		}
		delay := jitteredExpBackoff(failedAttempts, maxBackoff)
		log.Printf(
			"WARNING: failed %d attempts(s) to %s; will retry in %s: %s",
			failedAttempts,
			process,
			delay,
			err,
		)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
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
