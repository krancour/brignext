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
	// CleanerMaxFailures specifies the maximum number of consecutive times that
	// the cleanup process may fail before aborting the consumer.
	// Min: 0
	// Max: 10
	// Default: 2
	CleanerMaxFailures *uint8

	// HeartbeatInterval specifies how frequently to send out a heartbeat
	// indicating that the consumer is alive and functional.
	// Min: 5 seconds
	// Max: 5 minutes
	// Default: 30 seconds
	HeartbeatInterval *time.Duration
	// HeartbeatMaxFailures specifies the maximum number of consecutive times that
	// a heartbeat may fail to be sent before aborting.
	// Min: 0
	// Max: 10
	// Default: 2
	HeartbeatMaxFailures *uint8

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
	// ReceiverMaxFailures specifies the maximum number of consecutive times that
	// an attempt to retrieve a message from the global list of pending messages
	// may fail before aborting the consumer.
	// Min: 0
	// Max: 10
	// Default: 2
	ReceiverMaxFailures *uint8

	// SchedulerInterval specifies the interval at which messages with scheduled
	// handling times that have elapsed should ve transplanted to the global pending
	// messages list.
	// Min: 5 seconds
	// Max: 5 minutes
	// Default: 5 seconds
	SchedulerInterval *time.Duration
	// SchedulerMaxFailures specifies the maximum number of consecutive times that
	// the scheduler process may fail before aborting.
	// Min: 0
	// Max: 10
	// Default: 2
	SchedulerMaxFailures *uint8

	// ConcurrentHandlersCount specifies how many messages may be handed
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

	var maxCleanerMaxFailures uint8 = 10
	var defaultCleanerMaxFailures uint8 = 2
	if c.CleanerMaxFailures == nil {
		c.CleanerMaxFailures = &defaultCleanerMaxFailures
	} else if *c.CleanerMaxFailures > maxCleanerMaxFailures {
		c.CleanerMaxFailures = &maxCleanerMaxFailures
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

	var maxHeartbeatMaxFailures uint8 = 10
	var defaultHeartbeatMaxFailures uint8 = 2
	if c.HeartbeatMaxFailures == nil {
		c.HeartbeatMaxFailures = &defaultHeartbeatMaxFailures
	} else if *c.HeartbeatMaxFailures > maxHeartbeatMaxFailures {
		c.HeartbeatMaxFailures = &maxHeartbeatMaxFailures
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

	var maxReceiverMaxFailures uint8 = 10
	var defaultReceiverMaxFailures uint8 = 2
	if c.ReceiverMaxFailures == nil {
		c.ReceiverMaxFailures = &defaultReceiverMaxFailures
	} else if *c.ReceiverMaxFailures > maxReceiverMaxFailures {
		c.ReceiverMaxFailures = &maxReceiverMaxFailures
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

	var maxSchedulerMaxFailures uint8 = 10
	var defaultSchedulerMaxFailures uint8 = 2
	if c.SchedulerMaxFailures == nil {
		c.SchedulerMaxFailures = &defaultSchedulerMaxFailures
	} else if *c.SchedulerMaxFailures > maxSchedulerMaxFailures {
		c.SchedulerMaxFailures = &maxSchedulerMaxFailures
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
