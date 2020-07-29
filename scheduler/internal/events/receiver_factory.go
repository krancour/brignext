package events

import "context"

type ReceiverFactory interface {
	NewReceiver(projectID string) (Receiver, error)
	Close(context.Context) error
}
