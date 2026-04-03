package pkg

import (
	"errors"
	"fmt"
)

var (
	ErrUnprocessableEntity = errors.New("unprocessable entity")

	ErrConflict          = errors.New("conflict")
	ErrDuplicateTransfer = fmt.Errorf("%w: transfer already processed", ErrConflict)

	ErrNotFound                 = errors.New("not found")
	ErrWalletNotFound           = fmt.Errorf("%w: wallet not found", ErrNotFound)
	ErrWalletProjectionNotFound = fmt.Errorf("%w: wallet projection not found", ErrNotFound)

	ErrValidation              = errors.New("validation error")
	ErrAmountOverflow          = fmt.Errorf("%w: amount overflow", ErrValidation)
	ErrNegativeAmount          = fmt.Errorf("%w: amount must not be negative", ErrValidation)
	ErrNonPositiveAmount       = fmt.Errorf("%w: amount must be a positive non-zero integer", ErrValidation)
	ErrInsufficientBalance     = fmt.Errorf("%w: insufficient balance", ErrValidation)
	ErrBalanceLimitExceeded    = fmt.Errorf("%w: balance limit exceeded", ErrValidation)
	ErrMissingRequiredParam    = fmt.Errorf("%w: missing required parameter", ErrValidation)
	ErrInvalidUUID             = fmt.Errorf("%w: id must be a valid UUID", ErrValidation)
	ErrInvalidWalletID         = fmt.Errorf("%w: wallet_id must be a valid UUID", ErrValidation)
	ErrInvalidHolderID         = fmt.Errorf("%w: holder_id must be a valid UUID", ErrValidation)
	ErrInvalidTransferID       = fmt.Errorf("%w: transfer_id must be a valid UUID", ErrValidation)
	ErrInvalidToWalletID       = fmt.Errorf("%w: to_wallet_id must be a valid UUID", ErrValidation)
	ErrInvalidFromWalletID     = fmt.Errorf("%w: from_wallet_id must be a valid UUID", ErrValidation)
	ErrInvalidPaginationOffset = fmt.Errorf("%w: offset must be a non-negative integer", ErrValidation)
	ErrInvalidPaginationLimit  = fmt.Errorf("%w: limit must be a positive integer no greater than 100", ErrValidation)

	ErrInternal           = errors.New("internal server error")
	ErrUnknownEventType   = fmt.Errorf("%w: unknown event type", ErrInternal)
	ErrSubscriptionFailed = fmt.Errorf("%w: failed to subscribe", ErrInternal)
)

func IsPermanentError(err error) bool {
	return errors.Is(err, ErrValidation) || errors.Is(err, ErrUnknownEventType)
}
