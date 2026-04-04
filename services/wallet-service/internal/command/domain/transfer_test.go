package domain

import (
	"testing"
	"time"

	valueObject "wallet/wallet-service/pkg/domain/value_object"

	"github.com/stretchr/testify/assert"
)

func TestNewTransfer(t *testing.T) {
	t.Parallel()

	t.Run("It should create a transfer with the provided values", func(t *testing.T) {
		t.Parallel()

		id := valueObject.GenerateID()
		kind := TransferKindOutgoing
		amount := newMoney(t, 500)
		timestamp := time.Now()

		transfer := NewTransfer(id, kind, amount, CategoryFood, timestamp)

		assert.Equal(t, id.String(), transfer.ID().String())
		assert.Equal(t, TransferKindOutgoing, transfer.Kind())
		assert.Equal(t, CategoryFood, transfer.Category())
		assert.Equal(t, 500, transfer.Amount().Amount())
		assert.Equal(t, timestamp, transfer.Timestamp())
	})

	t.Run("It should create an incoming transfer", func(t *testing.T) {
		t.Parallel()

		transfer := NewTransfer(
			valueObject.GenerateID(),
			TransferKindIncoming,
			newMoney(t, 200),
			CategoryFuel,
			time.Now(),
		)

		assert.Equal(t, TransferKindIncoming, transfer.Kind())
		assert.Equal(t, CategoryFuel, transfer.Category())
	})
}

func TestNewOutgoingTransfer(t *testing.T) {
	t.Parallel()

	t.Run("It should create an outgoing transfer", func(t *testing.T) {
		t.Parallel()

		id := valueObject.GenerateID()
		amount := newMoney(t, 300)
		timestamp := time.Now()

		transfer := NewOutgoingTransfer(id, amount, CategoryEssentials, timestamp)

		assert.Equal(t, id.String(), transfer.ID().String())
		assert.Equal(t, TransferKindOutgoing, transfer.Kind())
		assert.Equal(t, CategoryEssentials, transfer.Category())
		assert.Equal(t, 300, transfer.Amount().Amount())
		assert.Equal(t, timestamp, transfer.Timestamp())
	})
}

func TestNewIncomingTransfer(t *testing.T) {
	t.Parallel()

	t.Run("It should create an incoming transfer", func(t *testing.T) {
		t.Parallel()

		id := valueObject.GenerateID()
		amount := newMoney(t, 150)
		timestamp := time.Now()

		transfer := NewIncomingTransfer(id, amount, CategoryEntertainment, timestamp)

		assert.Equal(t, id.String(), transfer.ID().String())
		assert.Equal(t, TransferKindIncoming, transfer.Kind())
		assert.Equal(t, CategoryEntertainment, transfer.Category())
		assert.Equal(t, 150, transfer.Amount().Amount())
		assert.Equal(t, timestamp, transfer.Timestamp())
	})
}

func TestIsValidCategory(t *testing.T) {
	t.Parallel()

	t.Run("It should return true for all valid categories", func(t *testing.T) {
		t.Parallel()

		validCategories := []string{
			CategoryFood,
			CategoryFuel,
			CategorySports,
			CategoryHealth,
			CategoryTravel,
			CategoryEssentials,
			CategoryEntertainment,
			CategoryUnclassified,
		}

		for _, category := range validCategories {
			assert.True(t, IsValidCategory(category), "expected %q to be valid", category)
		}
	})

	t.Run("It should return false for invalid categories", func(t *testing.T) {
		t.Parallel()

		invalidCategories := []string{"", "shopping", "FOOD", "Fuel", " essentials", "unclassified "}

		for _, category := range invalidCategories {
			assert.False(t, IsValidCategory(category), "expected %q to be invalid", category)
		}
	})
}
