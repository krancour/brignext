package messaging

import "context"

// Consumer is an interface for any component that can consume messages from
// a reliable queue and handle them using a user-specified handler function.
type Consumer interface {
	// Consumer causes the consumer to carry out all of its functions. It blocks
	// until a fatal error is encountered or the context passed to it has been
	// canceled. Consume always returns a non-nil error.
	Consume(context.Context, HandlerFn) error
}
