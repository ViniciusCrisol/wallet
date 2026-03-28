package pkg

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrValidation          = errors.New("validation")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrInternalServerError = errors.New("internal server error")

	ErrInvalidUUID          = fmt.Errorf("%w: id is not a valid uuid", ErrValidation)
	ErrInvalidAmount        = fmt.Errorf("%w: amount must be a positive integer", ErrValidation)
	ErrInsufficientBalance  = fmt.Errorf("%w: insufficient balance to complete the transfer", ErrValidation)
	ErrBalanceLimitExceeded = fmt.Errorf("%w: transfer would exceed the maximum allowed balance", ErrValidation)
	ErrWalletNotFound       = fmt.Errorf("%w: wallet not found", ErrNotFound)
	ErrUnknownEventType     = fmt.Errorf("%wunknown event type", ErrInternalServerError)
)
