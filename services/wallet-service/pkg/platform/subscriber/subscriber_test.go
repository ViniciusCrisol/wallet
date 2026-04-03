package subscriber

import (
	"errors"
	"reflect"
	"testing"
	"unsafe"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
)

func newKurrentDBError(code kurrentdb.ErrorCode) error {
	e := &kurrentdb.Error{}
	rv := reflect.ValueOf(e).Elem()
	field := rv.FieldByName("code")
	ptr := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr()))
	ptr.Elem().Set(reflect.ValueOf(code))
	return e
}

func TestIsKurrentDBNotFoundError(t *testing.T) {
	t.Parallel()

	t.Run("It should return false when the error is nil", func(t *testing.T) {
		t.Parallel()

		result := IsKurrentDBNotFoundError(nil)
		assert.False(t, result)
	})

	t.Run("It should return false when the error is a generic Go error", func(t *testing.T) {
		t.Parallel()

		err := errors.New("some unexpected error")
		result := IsKurrentDBNotFoundError(err)
		assert.False(t, result)
	})

	t.Run("It should return true when the error is a KurrentDB ResourceNotFound error", func(t *testing.T) {
		t.Parallel()

		err := newKurrentDBError(kurrentdb.ErrorCodeResourceNotFound)
		result := IsKurrentDBNotFoundError(err)
		assert.True(t, result)
	})

	t.Run("It should return false when the error is a KurrentDB error with a different error code", func(t *testing.T) {
		t.Parallel()

		err := newKurrentDBError(kurrentdb.ErrorCodeUnknown)
		result := IsKurrentDBNotFoundError(err)
		assert.False(t, result)
	})
}

func TestIsKurrentDBConcurrencyError(t *testing.T) {
	t.Parallel()

	t.Run("It should return false when the error is nil", func(t *testing.T) {
		t.Parallel()

		result := IsKurrentDBConcurrencyError(nil)
		assert.False(t, result)
	})

	t.Run("It should return false when the error is a generic Go error", func(t *testing.T) {
		t.Parallel()

		err := errors.New("some unexpected error")
		result := IsKurrentDBConcurrencyError(err)
		assert.False(t, result)
	})

	t.Run("It should return true when the error is a KurrentDB WrongExpectedVersion error", func(t *testing.T) {
		t.Parallel()

		err := newKurrentDBError(kurrentdb.ErrorCodeWrongExpectedVersion)
		result := IsKurrentDBConcurrencyError(err)
		assert.True(t, result)
	})

	t.Run("It should return false when the error is a KurrentDB error with a different error code", func(t *testing.T) {
		t.Parallel()

		err := newKurrentDBError(kurrentdb.ErrorCodeUnknown)
		result := IsKurrentDBConcurrencyError(err)
		assert.False(t, result)
	})
}
