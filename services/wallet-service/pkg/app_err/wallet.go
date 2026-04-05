package app_err

import "fmt"

var (
	ErrInsufficientBalance  = fmt.Errorf("%w: insufficient balance", ErrValidation)
	ErrBalanceLimitExceeded = fmt.Errorf("%w: balance limit exceeded", ErrValidation)

	ErrWalletNotFound           = fmt.Errorf("%w: wallet not found", ErrNotFound)
	ErrDuplicateTransfer        = fmt.Errorf("%w: transfer already processed", ErrConflict)
	ErrWalletProjectionNotFound = fmt.Errorf("%w: wallet projection not found", ErrNotFound)

	ErrAmountOverflow    = fmt.Errorf("%w: amount overflow", ErrValidation)
	ErrNegativeAmount    = fmt.Errorf("%w: amount must not be negative", ErrValidation)
	ErrNonPositiveAmount = fmt.Errorf("%w: amount must be a positive non-zero integer", ErrValidation)

	ErrInvalidWalletID      = fmt.Errorf("%w: wallet_id must be a valid UUID", ErrValidation)
	ErrInvalidHolderID      = fmt.Errorf("%w: holder_id must be a valid UUID", ErrValidation)
	ErrInvalidTransferID    = fmt.Errorf("%w: transfer_id must be a valid UUID", ErrValidation)
	ErrInvalidToWalletID    = fmt.Errorf("%w: to_wallet_id must be a valid UUID", ErrValidation)
	ErrInvalidFromWalletID  = fmt.Errorf("%w: from_wallet_id must be a valid UUID", ErrValidation)
	ErrInvalidCreatedAt     = fmt.Errorf("%w: created_at must be a valid RFC3339 timestamp", ErrValidation)
	ErrInvalidTransferredAt = fmt.Errorf("%w: transferred_at must be a valid RFC3339 timestamp", ErrValidation)
	ErrInvalidCategory      = fmt.Errorf("%w: category must be one of: food, fuel, sports, health, travel, essentials, entertainment, unclassified", ErrValidation)
)
