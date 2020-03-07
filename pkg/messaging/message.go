package messaging

import (
	"encoding/json"
	"time"

	uuid "github.com/satori/go.uuid"
)

// Message is an interface to be implemented by types that represent a single
// message.
type Message interface {
	// ID returns the unique message identifier.
	ID() string
	// Body returns the messsage body.
	Body() []byte
	// HandleTime returns the that the message indicates it should be handled at
	// or after, if any.
	HandleTime() *time.Time
	// ToJSON returns a []byte containing a JSON representation of the Message.
	ToJSON() ([]byte, error)
}

type message struct {
	IDAttr         string     `json:"id"`
	BodyAttr       []byte     `json:"body"`
	HandleTimeAttr *time.Time `json:"handleTime"`
}

// NewMessage returns a new Message.
func NewMessage(body []byte) Message {
	return &message{
		IDAttr:   uuid.NewV4().String(),
		BodyAttr: body,
	}
}

// NewScheduledMessage returns a new Message that should be handled at or after
// a specified time.
func NewScheduledMessage(
	body []byte,
	time time.Time,
) Message {
	mIface := NewMessage(body)
	m := mIface.(*message)
	m.HandleTimeAttr = &time
	return m
}

// NewDelayedMessage returns a new Message that should be handled after a
// specified duration.
func NewDelayedMessage(
	body []byte,
	delay time.Duration,
) Message {
	return NewScheduledMessage(body, time.Now().Add(delay))
}

// NewMessageFromJSON returns a new Message unmarshalled from the provided
// []byte.
func NewMessageFromJSON(jsonBytes []byte) (Message, error) {
	m := &message{}
	if err := json.Unmarshal(jsonBytes, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *message) ID() string {
	return m.IDAttr
}

func (m *message) Body() []byte {
	return m.BodyAttr
}

func (m *message) HandleTime() *time.Time {
	return m.HandleTimeAttr
}

func (m *message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}
