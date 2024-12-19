package event

import "time"

// BusConfig is the basic configuration structure for a [Bus] instance.
//
// Concrete implementations of the [Bus] interface might embed this structure to customize
// according to their own needs.
type BusConfig struct {
	ListenerTimeout      time.Duration
	ListenerInterceptors []ListenerInterceptorFunc
}

// BusOption is a routine signature enabling [Bus] instances to apply certain configurations through
// the "options" Go functional pattern.
type BusOption func(*BusConfig)

// WithListenerTimeout sets the timeout for a [ListenerFunc] routine execution.
func WithListenerTimeout(timeout time.Duration) BusOption {
	return func(cfg *BusConfig) {
		cfg.ListenerTimeout = timeout
	}
}

// WithListenerInterceptors sets the global slice of listener interceptors of a [Bus] entity.
func WithListenerInterceptors(interceptors ...ListenerInterceptorFunc) BusOption {
	return func(cfg *BusConfig) {
		cfg.ListenerInterceptors = append(cfg.ListenerInterceptors, interceptors...)
	}
}
