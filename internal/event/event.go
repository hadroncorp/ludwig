package event

// An Event is an occurrence that happened within a system.
type Event interface {
	// Topic the topic name this event is bound to.
	Topic() string
	// Message content (aka. payload) of the event in binary format.
	Message() []byte
}

// Option is a routine signature enabling [Event] instances to set certain fields through
// the "options" Go functional pattern.
//
//	NOTE: This signature can only be implemented by this package.
type Option func(*eventInternal)

// WithTopic sets the topic name of an [Event].
func WithTopic(topic string) Option {
	return func(e *eventInternal) {
		e.setTopic(topic)
	}
}

// WithMessage sets the content of an [Event].
func WithMessage(msg []byte) Option {
	return func(e *eventInternal) {
		e.setMessage(msg)
	}
}

// The default concrete implementation of [Event].
type eventInternal struct {
	topic string
	msg   []byte
}

var _ Event = (*eventInternal)(nil)

// NewEvent allocates a new [Event] instance.
func NewEvent(opts ...Option) Event {
	ev := &eventInternal{}
	for _, opt := range opts {
		opt(ev)
	}
	return ev
}

func (e *eventInternal) setTopic(topic string) {
	e.topic = topic
}

func (e *eventInternal) setMessage(msg []byte) {
	e.msg = msg
}

func (e *eventInternal) Topic() string {
	return e.topic
}

func (e *eventInternal) Message() []byte {
	return e.msg
}
