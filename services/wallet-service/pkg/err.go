package pkg

import (
	"errors"
	"fmt"
)

var (
	ErrInternal   = errors.New("")
	ErrNotFound   = errors.New("")
	ErrValidation = errors.New("")

	ErrUnknownEventType = fmt.Errorf("%wunknown event type", ErrInternal)

	ErrInvalidUUID          = fmt.Errorf("%wID is not a valid UUID", ErrValidation)
	ErrInvalidAmount        = fmt.Errorf("%wamount must be a positive integer", ErrValidation)
	ErrInsufficientBalance  = fmt.Errorf("%winsufficient balance to complete the transfer", ErrValidation)
	ErrBalanceLimitExceeded = fmt.Errorf("%wtransfer would exceed the maximum allowed balance", ErrValidation)
)
