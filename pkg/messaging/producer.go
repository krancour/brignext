package messaging

// Producer is an interface for any component that can publish messages to a
// reliable queue.
type Producer interface {
	// Publish enqueues a Message for asynchronous handling by a Consumer.
	Publish(Message) error
}
