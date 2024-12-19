package event

import "context"

type (
	// ListenerFunc is a routine executed by [Bus] instances when an [Event] is published.
	//
	// Take in consideration that the [Bus.Subscribe] routine MUST be called before publishing an [Event]
	// to ensure this routine gets executed.
	ListenerFunc func(ctx context.Context, event Event) error

	// ListenerInterceptorFunc is a routine executed by [Bus] instances before a [ListenerFunc] gets executed.
	//
	// This enables users to adhere logic without interfering with the original [ListenerFunc].
	//
	// Learn more by reading about "chain of responsibility" design pattern.
	ListenerInterceptorFunc func(next ListenerFunc) ListenerFunc
)
