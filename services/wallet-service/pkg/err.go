package pkg

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound   = errors.New(":")
	ErrValidation = errors.New(":")

	ErrInvalidID            = fmt.Errorf("%wID cannot be empty", ErrValidation)
	ErrInvalidUUID          = fmt.Errorf("%wID is not a valid UUID", ErrValidation)
	ErrInvalidAmount        = fmt.Errorf("%wamount must be a positive integer", ErrValidation)
	ErrInsufficientBalance  = fmt.Errorf("%winsufficient balance to complete the transfer", ErrValidation)
	ErrBalanceLimitExceeded = fmt.Errorf("%wtransfer would exceed the maximum allowed balance", ErrValidation)
)
