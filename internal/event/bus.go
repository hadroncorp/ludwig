package event

import (
	"context"
	"errors"
)

// Bus is a communication component for systems to propagate occurrences within themselves.
type Bus interface {
	// Publish propagates `event` to a set of subscribers (if any).
	Publish(ctx context.Context, event Event) error
	// Subscribe registers a [ListenerFunc] routine to the given `topic`. The given listener routine
	// will be executed everytime a new [Event] is published.
	//
	// Moreover, this routine has accepts a set of [ListenerInterceptorFunc] through a variadic argument.
	// These interceptors will be executed after the global set of interceptors (defined by
	// [BusConfig.ListenerInterceptors]).
	//
	// Finally, take into consideration that [Bus] processes might be concurrent; therefore,
	// the user MUST call Publish routine after a Subscribe routine call to ensure this subscription gets properly executed.
	Subscribe(ctx context.Context, topic string, listenFunc ListenerFunc, interceptors ...ListenerInterceptorFunc) error
}

var (
	// ErrBusClosed occurs if the [Bus] instance receives a call of one of its routines but the instance is already
	// closed or is currently closing.
	ErrBusClosed = errors.New("bus is closed")
	// ErrNoSubscribers occurs if the specified `topic` has no subscribers registered.
	ErrNoSubscribers = errors.New("no subscribers are registered")
)
