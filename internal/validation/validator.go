package validation

import "context"

// Validator is a system component for structure validations.
type Validator interface {
	// Validate validates the given structure.
	Validate(ctx context.Context, v any) error
}
