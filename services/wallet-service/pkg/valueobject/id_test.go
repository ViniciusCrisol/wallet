package valueobject

import (
	"testing"

	"wallet/wallet-service/pkg"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewID(t *testing.T) {
	t.Run("It should return a valid ID when value is a valid UUID", func(t *testing.T) {
		validUUID := uuid.NewString()
		id, err := NewID(validUUID)
		assert.NoError(t, err)
		assert.Equal(t, validUUID, id.ToString())
	})

	t.Run("It should return an error when value is an empty string", func(t *testing.T) {
		_, err := NewID("")
		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when value is not a valid UUID", func(t *testing.T) {
		_, err := NewID("not-a-uuid")
		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})
}

func TestGenerateID(t *testing.T) {
	t.Run("It should return an ID with a non-empty string value", func(t *testing.T) {
		id := GenerateID()
		assert.NotEmpty(t, id.ToString())
	})

	t.Run("It should return an ID whose value is a valid UUID", func(t *testing.T) {
		id := GenerateID()
		_, err := uuid.Parse(id.ToString())
		assert.NoError(t, err)
	})

	t.Run("It should return a different ID on each call", func(t *testing.T) {
		first := GenerateID()
		second := GenerateID()
		assert.NotEqual(t, first.ToString(), second.ToString())
	})
}

func TestID_ToString(t *testing.T) {
	t.Run("It should return the original UUID string when ID is created with NewID", func(t *testing.T) {
		validUUID := uuid.NewString()
		id, _ := NewID(validUUID)
		assert.Equal(t, validUUID, id.ToString())
	})
}
