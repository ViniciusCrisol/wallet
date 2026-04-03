package persistence

import (
	"context"
	"errors"
	"testing"
	"time"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/pkg"
	valueObject "wallet/wallet-service/pkg/domain/value_object"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalletKurrentDBESHandler_Save(t *testing.T) {
	t.Parallel()

	t.Run("It should save successfully when wallet has uncommitted events", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		wallet := newWallet(t)
		assert.NoError(t, esHandler.Save(context.Background(), wallet))
	})

	t.Run("It should return nil when wallet has no uncommitted events", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		wallet := newWallet(t)
		wallet.Commit()
		assert.NoError(t, esHandler.Save(context.Background(), wallet))
	})

	t.Run("It should save multiple events when wallet has multiple uncommitted events", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		wallet := newWallet(t)
		amount, err := valueObject.NewMoney(500)
		assert.NoError(t, err)
		assert.NoError(t, wallet.ReceiveFundsTransfer(domain.ReceiveFundsTransferCommand{
			Amount:       amount,
			TransferID:   valueObject.GenerateID(),
			FromWalletID: valueObject.GenerateID(),
			Timestamp:    time.Now(),
		}))
		assert.NoError(t, esHandler.Save(context.Background(), wallet))
	})

	t.Run("It should return an error when saving the same wallet stream twice with conflicting revisions", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		wallet := newWallet(t)
		assert.NoError(t, esHandler.Save(context.Background(), wallet))

		conflictWallet := domain.NewWallet(domain.CreateWalletCommand{
			WalletID:  wallet.ID(),
			HolderID:  valueObject.GenerateID(),
			Timestamp: time.Now(),
		})
		assert.True(t, errors.Is(esHandler.Save(context.Background(), conflictWallet), pkg.ErrConflict))
	})
}

func TestWalletKurrentDBESHandler_Find(t *testing.T) {
	t.Parallel()

	t.Run("It should return false when wallet stream does not exist", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		nonExistentID := valueObject.GenerateID()

		_, found, err := esHandler.Find(context.Background(), nonExistentID)

		assert.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("It should return the wallet and true when wallet was previously saved", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		wallet := newWallet(t)
		assert.NoError(t, esHandler.Save(context.Background(), wallet))

		result, found, err := esHandler.Find(context.Background(), wallet.ID())

		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, wallet.ID().String(), result.ID().String())
	})

	t.Run("It should reconstruct the correct balance when funds were received before saving", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		wallet := newWallet(t)
		amount, err := valueObject.NewMoney(750)
		assert.NoError(t, err)
		assert.NoError(t, wallet.ReceiveFundsTransfer(domain.ReceiveFundsTransferCommand{
			Amount:       amount,
			TransferID:   valueObject.GenerateID(),
			FromWalletID: valueObject.GenerateID(),
			Timestamp:    time.Now(),
		}))
		assert.NoError(t, esHandler.Save(context.Background(), wallet))

		result, found, err := esHandler.Find(context.Background(), wallet.ID())

		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, 750, result.Balance().Amount())
	})

	t.Run("It should reconstruct the correct balance after a receive and transfer cycle", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		wallet := newWallet(t)

		received, err := valueObject.NewMoney(1000)
		assert.NoError(t, err)
		assert.NoError(t, wallet.ReceiveFundsTransfer(domain.ReceiveFundsTransferCommand{
			Amount:       received,
			TransferID:   valueObject.GenerateID(),
			FromWalletID: valueObject.GenerateID(),
			Timestamp:    time.Now(),
		}))
		assert.NoError(t, esHandler.Save(context.Background(), wallet))

		reloaded, found, err := esHandler.Find(context.Background(), wallet.ID())
		assert.NoError(t, err)
		require.True(t, found)

		transferred, err := valueObject.NewMoney(400)
		assert.NoError(t, err)
		assert.NoError(t, reloaded.TransferFunds(domain.TransferFundsCommand{
			Amount:     transferred,
			TransferID: valueObject.GenerateID(),
			ToWalletID: valueObject.GenerateID(),
			Timestamp:  time.Now(),
		}))
		assert.NoError(t, esHandler.Save(context.Background(), reloaded))

		result, found, err := esHandler.Find(context.Background(), wallet.ID())
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, 600, result.Balance().Amount())
	})
}

func TestWalletKurrentDBESHandler_buildStreamName(t *testing.T) {
	t.Parallel()

	t.Run("It should return a stream name prefixed with 'wallet-' when given a valid ID", func(t *testing.T) {
		t.Parallel()

		id, _ := valueObject.NewID("550e8400-e29b-41d4-a716-446655440000")
		esHandler := &WalletKurrentDBESHandler{}
		result := esHandler.buildStreamName(id)
		assert.Equal(t, "wallet-550e8400-e29b-41d4-a716-446655440000", result)
	})
}
