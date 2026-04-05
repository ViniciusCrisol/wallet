package app_err

import (
	"errors"
	"fmt"
)

var (
	ErrConflict            = errors.New("conflict")
	ErrNotFound            = errors.New("not found")
	ErrUnprocessableEntity = errors.New("unprocessable entity")

	ErrInternal           = errors.New("internal server error")
	ErrUnknownEventType   = fmt.Errorf("%w: unknown event type", ErrInternal)
	ErrSubscriptionFailed = fmt.Errorf("%w: failed to subscribe", ErrInternal)

	ErrValidation              = errors.New("validation error")
	ErrInvalidUUID             = fmt.Errorf("%w: id must be a valid UUID", ErrValidation)
	ErrMissingRequiredParam    = fmt.Errorf("%w: missing required parameter", ErrValidation)
	ErrInvalidPaginationOffset = fmt.Errorf("%w: offset must be a positive integer", ErrValidation)
	ErrInvalidPaginationLimit  = fmt.Errorf("%w: limit must be a positive integer no greater than 100", ErrValidation)
)

func IsPermanentError(err error) bool {
	return errors.Is(err, ErrValidation) || errors.Is(err, ErrUnknownEventType)
}
