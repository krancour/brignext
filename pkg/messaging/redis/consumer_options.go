package redis

import "time"

// ConsumerOptions represents configutation options for a Consumer.
type ConsumerOptions struct {
	// RedisPrefix specifies a prefix for all Redis keys to effect some
	// rudimentary namespacing within a single Redis database.
	RedisPrefix string

	// CleanerInterval specifies how frequently to reclaim messages from dead
	// consumers.
	// Min: 5 seconds
	// Max: 5 minutes
	// Default: 1 minute
	CleanerInterval *time.Duration
	// DeadConsumerThreshold specifies how much time must elapse since its last
	// heartbeat for a consumer to be considered dead and in need of cleanup.
	// Min: 5 seconds
	// Max: 5 minutes
	// Default: 1 minute
	CleanerDeadConsumerThreshold *time.Duration
	// CleanerMaxAttempts specifies the maximum number of consecutive times that
	// the cleanup process may fail before aborting the consumer.
	// Min: 1
	// Max: 10
	// Default: 3
	CleanerMaxAttempts *uint8

	// HeartbeatInterval specifies how frequently to send out a heartbeat
	// indicating that the consumer is alive and functional.
	// Min: 5 seconds
	// Max: 5 minutes
	// Default: 30 seconds
	HeartbeatInterval *time.Duration
	// HeartbeatMaxAttempts specifies the maximum number of consecutive times that
	// a heartbeat may fail to be sent before aborting the consumer.
	// Min: 1
	// Max: 10
	// Default: 3
	HeartbeatMaxAttempts *uint8

	// ReceiverPauseInterval specifies the interval to pause before the next
	// attempt to retrieve a message from the global list of pending messages if
	// and only if the previous attempt retrieved nothing. This can be tuned to
	// achieve a balance between latency and the desire to not tax the CPU, the
	// network, or the database in situations where the global list of pending
	// messages is empty for a prolonged period.
	// Min: 1 second
	// Max: 1 minute
	// Default: 5 seconds
	ReceiverNoResultPauseInterval *time.Duration
	// ReceiverMaxAttempts specifies the maximum number of consecutive times that
	// an attempt to retrieve a message from the global list of pending messages
	// may fail before aborting the consumer.
	// Min: 1
	// Max: 10
	// Default: 3
	ReceiverMaxAttempts *uint8

	// SchedulerInterval specifies the interval at which messages with scheduled
	// handling times that have elapsed should ve transplanted to the global pending
	// messages list.
	// Min: 5 seconds
	// Max: 5 minutes
	// Default: 5 seconds
	SchedulerInterval *time.Duration
	// SchedulerMaxAttempts specifies the maximum number of consecutive times that
	// the scheduler process may fail before aborting.
	// Min: 1
	// Max: 10
	// Default: 3
	SchedulerMaxAttempts *uint8

	// ConcurrentReceiversCount specifies how many messages may be received
	// concurrently.
	// Min: 1
	// Max: 255
	// Default: 1
	ConcurrentReceiversCount *uint8

	// ConcurrentHandlersCount specifies how many messages may be handled
	// concurrently.
	// Min: 1
	// Max: 255
	// Default: 5
	ConcurrentHandlersCount *uint8

	// ShutdownGracePeriod specifies the maximum interval to wait for all of the
	// consumer's concurrently executing components to shut down gracefully before
	// the Run function returns control to the caller.
	// Min: 0
	// Max: none
	// Default: 10 seconds
	ShutdownGracePeriod *time.Duration
}

func (c *ConsumerOptions) applyDefaults() {
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

	minCleanerDeadConsumerThreshold := 5 * time.Second
	maxCleanerDeadConsumerThreshold := 5 * time.Minute
	defaultCleanerDeadConsumerThreshold := time.Minute
	if c.CleanerDeadConsumerThreshold == nil {
		c.CleanerDeadConsumerThreshold = &defaultCleanerDeadConsumerThreshold
	} else if *c.CleanerDeadConsumerThreshold < minCleanerDeadConsumerThreshold {
		c.CleanerDeadConsumerThreshold = &minCleanerDeadConsumerThreshold
	} else if *c.CleanerDeadConsumerThreshold > maxCleanerDeadConsumerThreshold {
		c.CleanerDeadConsumerThreshold = &maxCleanerDeadConsumerThreshold
	}

	var minCleanerMaxAttemps uint8 = 1
	var maxCleanerMaxAttemps uint8 = 10
	var defaultCleanerMaxAttempts uint8 = 3
	if c.CleanerMaxAttempts == nil {
		c.CleanerMaxAttempts = &defaultCleanerMaxAttempts
	} else if *c.CleanerMaxAttempts < minCleanerMaxAttemps {
		c.CleanerMaxAttempts = &minCleanerMaxAttemps
	} else if *c.CleanerMaxAttempts > maxCleanerMaxAttemps {
		c.CleanerMaxAttempts = &maxCleanerMaxAttemps
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

	var minHeartbeatMaxAttempts uint8 = 1
	var maxHeartbeatMaxAttempts uint8 = 10
	var defaultHeartbeatMaxAttempts uint8 = 3
	if c.HeartbeatMaxAttempts == nil {
		c.HeartbeatMaxAttempts = &defaultHeartbeatMaxAttempts
	} else if *c.HeartbeatMaxAttempts < minHeartbeatMaxAttempts {
		c.HeartbeatMaxAttempts = &minHeartbeatMaxAttempts
	} else if *c.HeartbeatMaxAttempts > maxHeartbeatMaxAttempts {
		c.HeartbeatMaxAttempts = &maxHeartbeatMaxAttempts
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

	var minReceiverMaxAttempts uint8 = 1
	var maxReceiverMaxAttempts uint8 = 10
	var defaultReceiverMaxAttempts uint8 = 3
	if c.ReceiverMaxAttempts == nil {
		c.ReceiverMaxAttempts = &defaultReceiverMaxAttempts
	} else if *c.ReceiverMaxAttempts < minReceiverMaxAttempts {
		c.ReceiverMaxAttempts = &minReceiverMaxAttempts
	} else if *c.ReceiverMaxAttempts > maxReceiverMaxAttempts {
		c.ReceiverMaxAttempts = &maxReceiverMaxAttempts
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

	var minSchedulerMaxAttempts uint8 = 1
	var maxSchedulerMaxAttempts uint8 = 10
	var defaultSchedulerMaxAttempts uint8 = 3
	if c.SchedulerMaxAttempts == nil {
		c.SchedulerMaxAttempts = &defaultSchedulerMaxAttempts
	} else if *c.SchedulerMaxAttempts < minSchedulerMaxAttempts {
		c.SchedulerMaxAttempts = &minSchedulerMaxAttempts
	} else if *c.SchedulerMaxAttempts > maxSchedulerMaxAttempts {
		c.SchedulerMaxAttempts = &maxSchedulerMaxAttempts
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
