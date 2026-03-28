package pkg

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrValidation          = errors.New("validation error")
	ErrConflict            = errors.New("conflict")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrInternalServerError = errors.New("internal server error")

	ErrInvalidUUID = fmt.Errorf("%w: id must be a valid UUID", ErrValidation)

	ErrWalletNotFound       = fmt.Errorf("%w: wallet not found", ErrNotFound)
	ErrInsufficientBalance  = fmt.Errorf("%w: insufficient balance", ErrValidation)
	ErrBalanceLimitExceeded = fmt.Errorf("%w: balance limit would be exceeded", ErrValidation)
	ErrNegativeOrZeroAmount = fmt.Errorf("%w: amount must be a positive non-zero integer", ErrValidation)

	ErrUnknownEventType = fmt.Errorf("%w: unknown event type", ErrInternalServerError)
)
