package apperr

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrValidation          = errors.New("validation error")
	ErrConflict            = errors.New("conflict")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrInternal            = errors.New("internal server error")

	ErrInvalidUUID          = fmt.Errorf("%w: id must be a valid UUID", ErrValidation)
	ErrMissingRequiredParam = fmt.Errorf("%w: missing required parameter", ErrValidation)

	ErrWalletNotFound           = fmt.Errorf("%w: wallet not found", ErrNotFound)
	ErrWalletProjectionNotFound = fmt.Errorf("%w: wallet projection not found", ErrNotFound)
	ErrInsufficientBalance      = fmt.Errorf("%w: insufficient balance", ErrValidation)
	ErrBalanceLimitExceeded     = fmt.Errorf("%w: balance limit would be exceeded", ErrValidation)
	ErrNegativeAmount           = fmt.Errorf("%w: amount must not be negative", ErrValidation)
	ErrNonPositiveAmount        = fmt.Errorf("%w: amount must be a positive non-zero integer", ErrValidation)
	ErrAmountOverflow           = fmt.Errorf("%w: amount overflow", ErrValidation)

	ErrUnknownEventType = fmt.Errorf("%w: unknown event type", ErrInternal)
)

func IsPermanentError(err error) bool {
	return errors.Is(err, ErrValidation) || errors.Is(err, ErrUnknownEventType)
}
