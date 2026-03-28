package pkg

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound            = errors.New("")
	ErrValidation          = errors.New("")
	ErrConflict            = errors.New("conflict")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrInternalServerError = errors.New("internal server error")

	ErrInvalidUUID = fmt.Errorf("%wid must be a valid UUID", ErrValidation)

	ErrWalletNotFound       = fmt.Errorf("%wwallet not found", ErrNotFound)
	ErrInsufficientBalance  = fmt.Errorf("%winsufficient balance", ErrValidation)
	ErrBalanceLimitExceeded = fmt.Errorf("%wbalance limit would be exceeded", ErrValidation)
	ErrNegativeOrZeroAmount = fmt.Errorf("%wamount must be a positive non-zero integer", ErrValidation)

	ErrUnknownEventType = fmt.Errorf("%wunknown event type", ErrInternalServerError)
)
