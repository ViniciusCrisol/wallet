package uuid

import (
	"testing"

	googleuuid "github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

func TestIsValid(t *testing.T) {
	t.Parallel()

	t.Run("It should return true for a valid UUID v4", func(t *testing.T) {
		t.Parallel()

		assert.True(t, IsValid(googleuuid.New().String()))
	})

	t.Run("It should return true for a valid UUID v7", func(t *testing.T) {
		t.Parallel()

		v7, _ := googleuuid.NewV7()
		assert.True(t, IsValid(v7.String()))
	})

	t.Run("It should return false for an empty string", func(t *testing.T) {
		t.Parallel()

		assert.False(t, IsValid(""))
	})

	t.Run("It should return false for a non-UUID string", func(t *testing.T) {
		t.Parallel()

		assert.False(t, IsValid("not-a-uuid"))
	})

	t.Run("It should return false for a UUID with incorrect format", func(t *testing.T) {
		t.Parallel()

		assert.False(t, IsValid("550e8400-e29b-41d4-a716"))
	})
}

func TestNewUUID(t *testing.T) {
	t.Parallel()

	t.Run("It should return a non-empty string", func(t *testing.T) {
		t.Parallel()

		result := NewUUID()
		assert.NotEmpty(t, result)
	})

	t.Run("It should return a valid UUID string", func(t *testing.T) {
		t.Parallel()

		result := NewUUID()
		_, err := googleuuid.Parse(result)
		assert.NoError(t, err)
	})

	t.Run("It should return a different value on each call", func(t *testing.T) {
		t.Parallel()

		first := NewUUID()
		second := NewUUID()
		assert.NotEqual(t, first, second)
	})
}
