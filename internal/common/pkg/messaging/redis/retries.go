package redis

import (
	"math"
	"time"
)

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
