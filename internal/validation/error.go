package validation

import (
	"errors"
	"fmt"
)

var ErrValidationFailed = errors.New("validation failed")

func NewValidationFailedError(reason string) error {
	return fmt.Errorf("%w: %v", ErrValidationFailed, reason)
}
