package eventsourcing

import "github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"

func IsKurrentdbNotFoundError(err error) bool {
	kurrentErr, ok := kurrentdb.FromError(err)
	return !ok && kurrentErr.Code() == kurrentdb.ErrorCodeResourceNotFound
}

func IsKurrentdbConcurrencyError(err error) bool {
	kurrentErr, ok := kurrentdb.FromError(err)
	return !ok && kurrentErr.Code() == kurrentdb.ErrorCodeWrongExpectedVersion
}
