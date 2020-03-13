package redis

import "time"

// ConsumerOptions represents configutation options for a Consumer.
type ConsumerOptions struct {
	// RedisPrefix specifies a prefix for all Redis keys to effect some
	// rudimentary namespacing within a single Redis database.
	RedisPrefix string

	// RedisOperationMaxAttempts specifies the maximum number of consecutive times
	// that any discrete Redis operation may fail before aborting the consumer.
	//
	// Min: 1
	// Max: 10
	// Default: 3
	RedisOperationMaxAttempts *uint8

	// RedisOperationMaxBackoff specifies a cap on the exponentially increasing
	// delay before re-attempting any dsicrete Redis operation that has previously
	// failed.
	//
	// Min: 10 seconds
	// Max: 10 minutes
	// Default: 1 minute
	RedisOperationMaxBackoff *time.Duration

	// LoneConsumer specifies whether the creator of the consumer is offering a
	// STRONG GUARANTEE that this consumer will NEVER run concurrently with
	// another consumer of the same queue.
	//
	// This is useful for cases where messages must be handled in the order they
	// were received. A lone consumer that can assume no other consumers run
	// concurrently with itself can, during initialization, eagerly and
	// synchronously reclaim ALL messages previously claimed by other (dead)
	// consumers, such that a dead consumer's incomplete work is resumed prior to
	// taking on new work. To further illustrate, this could be useful when using
	// a queue to govern/limit concurrent work in some resource-constrained
	// backend system.
	LoneConsumer bool

	// CleanerInterval specifies how frequently to reclaim messages from dead
	// consumers. This setting is not used by "lone consumers."
	//
	// Min: 5 seconds
	// Max: 5 minutes
	// Default: 1 minute
	CleanerInterval *time.Duration

	// HeartbeatInterval specifies how frequently to send out a heartbeat
	// indicating that the consumer is alive and functional. This setting is not
	// used by "lone consumers."
	//
	// Min: 5 seconds
	// Max: 5 minutes
	// Default: 30 seconds
	HeartbeatInterval *time.Duration

	// ReceiverPauseInterval specifies the interval to pause before the next
	// attempt to retrieve a message from the global list of pending messages if
	// and only if the previous attempt retrieved nothing. This can be tuned to
	// achieve a balance between latency and the desire to not tax the CPU, the
	// network, or the database in situations where the global list of pending
	// messages is empty for a prolonged period.
	//
	// Min: 1 second
	// Max: 1 minute
	// Default: 5 seconds
	ReceiverNoResultPauseInterval *time.Duration

	// SchedulerInterval specifies the interval at which messages with scheduled
	// handling times that have elapsed should ve transplanted to the global
	// pending messages list.
	//
	// Min: 5 seconds
	// Max: 5 minutes
	// Default: 5 seconds
	SchedulerInterval *time.Duration

	// ConcurrentReceiversCount specifies how many messages may be received
	// concurrently. This number can be adjusted with respect to the number of
	// concurrent handlers and the number of connections in the Redis client's
	// connection pool in order to either maximize throughput or conserve
	// resources.
	//
	// Receivers can become a bottleneck if outnumbered by many handlers that each
	// process a single message very quickly, as many handlers may then sit idle
	// while waiting for a receiver to pull a messages from the queue. To increase
	// throughput, consider increasing the number of receivers as high as the
	// number of handlers. There is no reason to increase the number of handlers
	// beyond this because some will sit idle at any given moment while waiting
	// for an available worker. Throughput can be further maximized by using a
	// Redis client with a connection pool large enough to accommodate all
	// unblocked goroutines concurrently. This number will be the larger of the
	// receiver count and the handler count.
	//
	// For scenarios where any number of workers each process a single message
	// more slowly, receivers will often sit idle waiting for an available worker.
	// In such scenarios where receivers are not the bottleneck, resources can be
	// conserved by decreasing the number of receivers as low as 1.
	//
	// Min: 1
	// Max: 255
	// Default: 1
	ConcurrentReceiversCount *uint8

	// ConcurrentHandlersCount specifies how many messages may be handled
	// concurrently. This number can be adjusted up to improve throughput or down
	// as a means of governing/limiting concurrent work in some
	// resource-constrained backend system.
	//
	// Min: 1
	// Max: 255
	// Default: 5
	ConcurrentHandlersCount *uint8

	// ShutdownGracePeriod specifies the maximum interval to wait for all of the
	// consumer's concurrently executing components to shut down gracefully before
	// the Run function returns control to the caller.
	//
	// Min: 0
	// Max: none
	// Default: 10 seconds
	ShutdownGracePeriod *time.Duration
}

func (c *ConsumerOptions) applyDefaults() {
	var minRedisOperationMaxAttemps uint8 = 1
	var maxRedisOperationMaxAttemps uint8 = 10
	var defaultRedisOperationMaxAttemps uint8 = 3
	if c.RedisOperationMaxAttempts == nil {
		c.RedisOperationMaxAttempts = &defaultRedisOperationMaxAttemps
	} else if *c.RedisOperationMaxAttempts < minRedisOperationMaxAttemps {
		c.RedisOperationMaxAttempts = &minRedisOperationMaxAttemps
	} else if *c.RedisOperationMaxAttempts > maxRedisOperationMaxAttemps {
		c.RedisOperationMaxAttempts = &maxRedisOperationMaxAttemps
	}

	var minRedisOperationMaxBackoff = 10 * time.Second
	var maxRedisOperationMaxBackoff = 10 * time.Minute
	var defaultRedisOperationMaxBackoff = time.Minute
	if c.RedisOperationMaxBackoff == nil {
		c.RedisOperationMaxBackoff = &defaultRedisOperationMaxBackoff
	} else if *c.RedisOperationMaxBackoff < minRedisOperationMaxBackoff {
		c.RedisOperationMaxBackoff = &minRedisOperationMaxBackoff
	} else if *c.RedisOperationMaxBackoff > maxRedisOperationMaxBackoff {
		c.RedisOperationMaxBackoff = &maxRedisOperationMaxBackoff
	}

	minCleanerInterval := 5 * time.Second
	maxCleanerInterval := 5 * time.Minute
	defaultCleanerInterval := time.Minute
	if c.CleanerInterval == nil {
		c.CleanerInterval = &defaultCleanerInterval
	} else if *c.CleanerInterval < minCleanerInterval {
		c.CleanerInterval = &minCleanerInterval
	} else if *c.CleanerInterval > maxCleanerInterval {
		c.CleanerInterval = &maxCleanerInterval
	}

	minHeartbeatInterval := 5 * time.Second
	maxHeartbeatInterval := 5 * time.Minute
	defaultHeartbeatInterval := 30 * time.Second
	if c.HeartbeatInterval == nil {
		c.HeartbeatInterval = &defaultHeartbeatInterval
	} else if *c.HeartbeatInterval < minHeartbeatInterval {
		c.HeartbeatInterval = &minHeartbeatInterval
	} else if *c.HeartbeatInterval > maxHeartbeatInterval {
		c.HeartbeatInterval = &maxHeartbeatInterval
	}

	minReceiverNoResultPauseInterval := time.Second
	maxReceiverNoResultPauseInterval := time.Minute
	defaultReceiverNoResultPauseInterval := 5 * time.Second
	if c.ReceiverNoResultPauseInterval == nil {
		c.ReceiverNoResultPauseInterval = &defaultReceiverNoResultPauseInterval
	} else if *c.ReceiverNoResultPauseInterval < minReceiverNoResultPauseInterval { // nolint: lll
		c.ReceiverNoResultPauseInterval = &minReceiverNoResultPauseInterval
	} else if *c.ReceiverNoResultPauseInterval > maxReceiverNoResultPauseInterval { // nolint: lll
		c.ReceiverNoResultPauseInterval = &maxReceiverNoResultPauseInterval
	}

	minSchedulerInterval := 5 * time.Second
	maxSchedulerInterval := 5 * time.Minute
	defaultSchedulerInterval := 5 * time.Second
	if c.SchedulerInterval == nil {
		c.SchedulerInterval = &defaultSchedulerInterval
	} else if *c.SchedulerInterval < minSchedulerInterval {
		c.SchedulerInterval = &minSchedulerInterval
	} else if *c.SchedulerInterval > maxSchedulerInterval {
		c.SchedulerInterval = &maxSchedulerInterval
	}

	var minConcurrentReceiversCount uint8 = 1
	var defaultConcurrentReceiversCount uint8 = 5
	if c.ConcurrentReceiversCount == nil {
		c.ConcurrentReceiversCount = &defaultConcurrentReceiversCount
	} else if *c.ConcurrentReceiversCount < minConcurrentReceiversCount {
		c.ConcurrentReceiversCount = &minConcurrentReceiversCount
	}

	var minConcurrentHandlersCount uint8 = 1
	var defaultConcurrentHandlersCount uint8 = 5
	if c.ConcurrentHandlersCount == nil {
		c.ConcurrentHandlersCount = &defaultConcurrentHandlersCount
	} else if *c.ConcurrentHandlersCount < minConcurrentHandlersCount {
		c.ConcurrentHandlersCount = &minConcurrentHandlersCount
	}

	var minShutdownGracePeriod time.Duration
	defaultShutdownGracePeriod := 10 * time.Second
	if c.ShutdownGracePeriod == nil {
		c.ShutdownGracePeriod = &defaultShutdownGracePeriod
	} else if *c.ShutdownGracePeriod < minShutdownGracePeriod {
		c.ShutdownGracePeriod = &minShutdownGracePeriod
	}
}
