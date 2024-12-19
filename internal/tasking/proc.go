package tasking

import "context"

// Starter defines a starting routine. Useful for processes that require to start a long-running background process.
type Starter interface {
	// Start starts the specific process.
	Start(ctx context.Context) error
}

// Shutdowner defines a stopping routine. Useful for processes that require to shut down a long-running background process.
type Shutdowner interface {
	// Shutdown stops the specific process.
	Shutdown(ctx context.Context) error
}
