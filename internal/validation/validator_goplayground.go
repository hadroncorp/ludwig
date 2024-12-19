package validation

import (
	"context"

	"github.com/go-playground/validator/v10"
)

// ValidatorGoPlayground is the concrete implementation of [Validator] interface.
//
// Uses [github.com/go-playground/validator/v10] library for internal operations.
type ValidatorGoPlayground struct {
	validator *validator.Validate
}

var _ Validator = (*ValidatorGoPlayground)(nil)

// NewGoPlayground allocates a new [ValidatorGoPlayground] instance.
func NewGoPlayground() ValidatorGoPlayground {
	return ValidatorGoPlayground{
		validator: validator.New(),
	}
}

func (p ValidatorGoPlayground) Validate(ctx context.Context, v any) error {
	// TODO: Add adapters. Convert go-playground errors to internal system errors.
	return p.validator.StructCtx(ctx, v)
}
