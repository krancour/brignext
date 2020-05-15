package messaging

import "context"

// Consumer is an interface for any component that can consume messages from
// a reliable queue and handle them using a user-specified handler function.
type Consumer interface {
	// Run causes the consumer to carry out all of its functions. It blocks until
	// a fatal error is encountered or the context passed to it has been canceled.
	Run(context.Context) error
}
