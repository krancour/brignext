package events

type AsyncEvent struct {
	EventID string
	Ack     func() error
}
