package events

import "context"

type SenderFactory interface {
	NewSender(projectID string) (Sender, error)
	Close(context.Context) error
}
