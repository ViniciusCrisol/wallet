package pkg

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewUUID(t *testing.T) {
	t.Run("It should return a non-empty string", func(t *testing.T) {
		result := NewUUID()
		assert.NotEmpty(t, result)
	})

	t.Run("It should return a valid UUID string", func(t *testing.T) {
		result := NewUUID()
		_, err := uuid.Parse(result)
		assert.NoError(t, err)
	})

	t.Run("It should return a different value on each call", func(t *testing.T) {
		first := NewUUID()
		second := NewUUID()
		assert.NotEqual(t, first, second)
	})
}
