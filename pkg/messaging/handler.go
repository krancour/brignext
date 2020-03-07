package messaging

import "context"

// HandlerFn is the signature for functions that consumers can call to
// handle a message
type HandlerFn func(context.Context, Message) error
