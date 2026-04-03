package pkg

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPermanentError(t *testing.T) {
	t.Run("It should return true when error is ErrValidation", func(t *testing.T) {
		result := IsPermanentError(ErrValidation)
		assert.True(t, result)
	})

	t.Run("It should return true when error is ErrUnknownEventType", func(t *testing.T) {
		result := IsPermanentError(ErrUnknownEventType)
		assert.True(t, result)
	})

	t.Run("It should return true when error is wrapped under ErrValidation", func(t *testing.T) {
		result := IsPermanentError(ErrNegativeAmount)
		assert.True(t, result)
	})

	t.Run("It should return true when error is any other wrapped ErrValidation", func(t *testing.T) {
		result := IsPermanentError(ErrNonPositiveAmount)
		assert.True(t, result)
	})

	t.Run("It should return false when error is ErrNotFound", func(t *testing.T) {
		result := IsPermanentError(ErrNotFound)
		assert.False(t, result)
	})

	t.Run("It should return false when error is ErrConflict", func(t *testing.T) {
		result := IsPermanentError(ErrConflict)
		assert.False(t, result)
	})

	t.Run("It should return false when error is nil", func(t *testing.T) {
		result := IsPermanentError(nil)
		assert.False(t, result)
	})

	t.Run("It should return false when error is a generic error", func(t *testing.T) {
		result := IsPermanentError(errors.New("some generic error"))
		assert.False(t, result)
	})
}
