package persistence

import (
	"errors"
	"testing"
	"time"

	"wallet/wallet-service/config"
	"wallet/wallet-service/internal/domain"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestESHandler(t *testing.T) *WalletKurrentESHandler {
	t.Helper()

	settings, err := kurrentdb.ParseConnectionString(config.Load().KurrentDBConnectionString)
	if err != nil {
		t.Fatalf("failed to parse connection string: %v", err)
	}
	db, err := kurrentdb.NewClient(settings)
	if err != nil {
		t.Fatalf("failed to create kurrentdb client: %v", err)
	}
	return NewWalletKurrentESHandler(db)
}

func newWallet(t *testing.T) domain.Wallet {
	t.Helper()

	return domain.NewWallet(domain.CreateWalletCommand{
		WalletID:  valueobject.GenerateID(),
		HolderID:  valueobject.GenerateID(),
		Timestamp: time.Now(),
	})
}

func TestWalletKurrentESHandler_Save(t *testing.T) {
	t.Parallel()

	t.Run("It should save successfully when wallet has uncommitted events", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		wallet := newWallet(t)
		assert.NoError(t, esHandler.Save(wallet))
	})

	t.Run("It should return nil when wallet has no uncommitted events", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		wallet := newWallet(t)
		wallet.Commit()
		assert.NoError(t, esHandler.Save(wallet))
	})

	t.Run("It should save multiple events when wallet has multiple uncommitted events", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		wallet := newWallet(t)
		amount, err := valueobject.NewMoney(500)
		assert.NoError(t, err)
		assert.NoError(t, wallet.ReceiveFundsTransfer(domain.ReceiveFundsTransferCommand{
			Amount:       amount,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		}))
		assert.NoError(t, esHandler.Save(wallet))
	})

	t.Run("It should return an error when saving the same wallet stream twice with conflicting revisions", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		wallet := newWallet(t)
		assert.NoError(t, esHandler.Save(wallet))

		conflictWallet := domain.NewWallet(domain.CreateWalletCommand{
			WalletID:  wallet.GetID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})
		assert.True(t, errors.Is(esHandler.Save(conflictWallet), pkg.ErrConflict))
	})
}

func TestWalletKurrentESHandler_Find(t *testing.T) {
	t.Parallel()

	t.Run("It should return false when wallet stream does not exist", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		nonExistentID := valueobject.GenerateID()

		_, found, err := esHandler.Find(nonExistentID)

		assert.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("It should return the wallet and true when wallet was previously saved", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		wallet := newWallet(t)
		assert.NoError(t, esHandler.Save(wallet))

		result, found, err := esHandler.Find(wallet.GetID())

		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, wallet.GetID().ToString(), result.GetID().ToString())
	})

	t.Run("It should reconstruct the correct balance when funds were received before saving", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		wallet := newWallet(t)
		amount, err := valueobject.NewMoney(750)
		assert.NoError(t, err)
		assert.NoError(t, wallet.ReceiveFundsTransfer(domain.ReceiveFundsTransferCommand{
			Amount:       amount,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		}))
		assert.NoError(t, esHandler.Save(wallet))

		result, found, err := esHandler.Find(wallet.GetID())

		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, 750, result.GetBalance().GetAmount())
	})

	t.Run("It should reconstruct the correct balance after a receive and transfer cycle", func(t *testing.T) {
		t.Parallel()

		esHandler := newTestESHandler(t)
		wallet := newWallet(t)

		received, err := valueobject.NewMoney(1000)
		assert.NoError(t, err)
		assert.NoError(t, wallet.ReceiveFundsTransfer(domain.ReceiveFundsTransferCommand{
			Amount:       received,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		}))
		assert.NoError(t, esHandler.Save(wallet))

		reloaded, found, err := esHandler.Find(wallet.GetID())
		assert.NoError(t, err)
		require.True(t, found)

		transferred, err := valueobject.NewMoney(400)
		assert.NoError(t, err)
		assert.NoError(t, reloaded.TransferFunds(domain.TransferFundsCommand{
			Amount:     transferred,
			TransferID: valueobject.GenerateID(),
			ToWalletID: valueobject.GenerateID(),
			Timestamp:  time.Now(),
		}))
		assert.NoError(t, esHandler.Save(reloaded))

		result, found, err := esHandler.Find(wallet.GetID())
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, 600, result.GetBalance().GetAmount())
	})
}

func TestWalletKurrentESHandler_buildStreamName(t *testing.T) {
	t.Run("It should return a stream name prefixed with 'wallet-' when given a valid ID", func(t *testing.T) {
		id, _ := valueobject.NewID("550e8400-e29b-41d4-a716-446655440000")
		esHandler := &WalletKurrentESHandler{}
		result := esHandler.buildStreamName(id)
		assert.Equal(t, "wallet-550e8400-e29b-41d4-a716-446655440000", result)
	})
}
