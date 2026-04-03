package valueobject

import (
	"testing"

	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/platform/uuid"

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
		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when value is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		_, err := NewID("not-a-uuid")
		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
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

func TestID_Equals(t *testing.T) {
	t.Parallel()

	t.Run("It should return true when both IDs have the same value", func(t *testing.T) {
		t.Parallel()

		value := uuid.NewUUID()
		id1, _ := NewID(value)
		id2, _ := NewID(value)
		assert.True(t, id1.Equals(id2))
	})

	t.Run("It should return false when IDs have different values", func(t *testing.T) {
		t.Parallel()

		id1 := GenerateID()
		id2 := GenerateID()
		assert.False(t, id1.Equals(id2))
	})

	t.Run("It should return true when comparing an ID to itself", func(t *testing.T) {
		t.Parallel()

		id := GenerateID()
		assert.True(t, id.Equals(id))
	})
}
