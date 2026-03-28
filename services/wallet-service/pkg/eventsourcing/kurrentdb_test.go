package eventsourcing

import (
	"errors"
	"reflect"
	"testing"
	"unsafe"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
)

func newKurrentdbError(code kurrentdb.ErrorCode) error {
	e := &kurrentdb.Error{}
	rv := reflect.ValueOf(e).Elem()
	field := rv.FieldByName("code")
	ptr := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr()))
	ptr.Elem().Set(reflect.ValueOf(code))
	return e
}

func TestIsKurrentdbNotFoundError(t *testing.T) {
	t.Run("It should return false when the error is nil", func(t *testing.T) {
		result := IsKurrentdbNotFoundError(nil)
		assert.False(t, result)
	})

	t.Run("It should return false when the error is a generic Go error", func(t *testing.T) {
		err := errors.New("some unexpected error")
		result := IsKurrentdbNotFoundError(err)
		assert.False(t, result)
	})

	t.Run("It should return true when the error is a KurrentDB ResourceNotFound error", func(t *testing.T) {
		err := newKurrentdbError(kurrentdb.ErrorCodeResourceNotFound)
		result := IsKurrentdbNotFoundError(err)
		assert.True(t, result)
	})

	t.Run("It should return false when the error is a KurrentDB error with a different error code", func(t *testing.T) {
		err := newKurrentdbError(kurrentdb.ErrorCodeUnknown)
		result := IsKurrentdbNotFoundError(err)
		assert.False(t, result)
	})
}

func TestIsKurrentdbConcurrencyError(t *testing.T) {
	t.Run("It should return false when the error is nil", func(t *testing.T) {
		result := IsKurrentdbConcurrencyError(nil)
		assert.False(t, result)
	})

	t.Run("It should return false when the error is a generic Go error", func(t *testing.T) {
		err := errors.New("some unexpected error")
		result := IsKurrentdbConcurrencyError(err)
		assert.False(t, result)
	})

	t.Run("It should return true when the error is a KurrentDB WrongExpectedVersion error", func(t *testing.T) {
		err := newKurrentdbError(kurrentdb.ErrorCodeWrongExpectedVersion)
		result := IsKurrentdbConcurrencyError(err)
		assert.True(t, result)
	})

	t.Run("It should return false when the error is a KurrentDB error with a different error code", func(t *testing.T) {
		err := newKurrentdbError(kurrentdb.ErrorCodeUnknown)
		result := IsKurrentdbConcurrencyError(err)
		assert.False(t, result)
	})
}
