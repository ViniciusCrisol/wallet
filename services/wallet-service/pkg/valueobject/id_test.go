package valueobject

import (
	"testing"

	"wallet/wallet-service/pkg/apperr"
	"wallet/wallet-service/pkg/uuid"

	"github.com/stretchr/testify/assert"
)

func TestNewID(t *testing.T) {
	t.Parallel()

	t.Run("It should return a valid ID when value is a valid UUID", func(t *testing.T) {
		t.Parallel()

		validUUID := uuid.NewUUID()
		id, err := NewID(validUUID)
		assert.NoError(t, err)
		assert.Equal(t, validUUID, id.String())
	})

	t.Run("It should return an error when value is an empty string", func(t *testing.T) {
		t.Parallel()

		_, err := NewID("")
		assert.ErrorIs(t, err, apperr.ErrInvalidUUID)
	})

	t.Run("It should return an error when value is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		_, err := NewID("not-a-uuid")
		assert.ErrorIs(t, err, apperr.ErrInvalidUUID)
	})
}

func TestGenerateID(t *testing.T) {
	t.Parallel()

	t.Run("It should return an ID with a non-empty string value", func(t *testing.T) {
		t.Parallel()

		id := GenerateID()
		assert.NotEmpty(t, id.String())
	})

	t.Run("It should return an ID whose value is a valid UUID", func(t *testing.T) {
		t.Parallel()

		id := GenerateID()
		assert.True(t, uuid.IsValid(id.String()))
	})

	t.Run("It should return a different ID on each call", func(t *testing.T) {
		t.Parallel()

		first := GenerateID()
		second := GenerateID()
		assert.NotEqual(t, first.String(), second.String())
	})
}

func TestID_String(t *testing.T) {
	t.Parallel()

	t.Run("It should return the original UUID string when ID is created with NewID", func(t *testing.T) {
		t.Parallel()

		validUUID := uuid.NewUUID()
		id, _ := NewID(validUUID)
		assert.Equal(t, validUUID, id.String())
	})
}
