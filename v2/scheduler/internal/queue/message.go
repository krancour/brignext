package queue

type Message struct {
	Message string
	Ack     func() error
}
