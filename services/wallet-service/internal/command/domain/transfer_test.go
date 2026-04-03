package domain

import (
	"testing"
	"time"

	"wallet/wallet-service/pkg/valueobject"

	"github.com/stretchr/testify/assert"
)

func TestNewTransfer(t *testing.T) {
	t.Parallel()

	t.Run("It should create a transfer with the provided values", func(t *testing.T) {
		t.Parallel()

		id := valueobject.GenerateID()
		kind := TransferKindOutgoing
		amount := newMoney(t, 500)
		timestamp := time.Now()

		transfer := NewTransfer(id, kind, amount, timestamp)

		assert.Equal(t, id.String(), transfer.ID().String())
		assert.Equal(t, TransferKindOutgoing, transfer.Kind())
		assert.Equal(t, 500, transfer.Amount().Amount())
		assert.Equal(t, timestamp, transfer.Timestamp())
	})

	t.Run("It should create an incoming transfer", func(t *testing.T) {
		t.Parallel()

		transfer := NewTransfer(
			valueobject.GenerateID(),
			TransferKindIncoming,
			newMoney(t, 200),
			time.Now(),
		)

		assert.Equal(t, TransferKindIncoming, transfer.Kind())
	})
}

func TestNewOutgoingTransfer(t *testing.T) {
	t.Parallel()

	t.Run("It should create an outgoing transfer", func(t *testing.T) {
		t.Parallel()

		id := valueobject.GenerateID()
		amount := newMoney(t, 300)
		timestamp := time.Now()

		transfer := NewOutgoingTransfer(id, amount, timestamp)

		assert.Equal(t, id.String(), transfer.ID().String())
		assert.Equal(t, TransferKindOutgoing, transfer.Kind())
		assert.Equal(t, 300, transfer.Amount().Amount())
		assert.Equal(t, timestamp, transfer.Timestamp())
	})
}

func TestNewIncomingTransfer(t *testing.T) {
	t.Parallel()

	t.Run("It should create an incoming transfer", func(t *testing.T) {
		t.Parallel()

		id := valueobject.GenerateID()
		amount := newMoney(t, 150)
		timestamp := time.Now()

		transfer := NewIncomingTransfer(id, amount, timestamp)

		assert.Equal(t, id.String(), transfer.ID().String())
		assert.Equal(t, TransferKindIncoming, transfer.Kind())
		assert.Equal(t, 150, transfer.Amount().Amount())
		assert.Equal(t, timestamp, transfer.Timestamp())
	})
}
