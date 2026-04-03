package postgresqlprojectordao

import (
	"context"
	"testing"
	"time"

	"wallet/wallet-service/pkg/platform/integrationevent"
	"wallet/wallet-service/pkg/platform/uuid"

	"github.com/stretchr/testify/assert"
)

func TestWalletPostgreSQLProjectionDAO_CreateWallet(t *testing.T) {
	t.Parallel()

	t.Run("It should successfully create a wallet when valid event is provided", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletPostgreSQLProjectionDAO(postgreSQLDB)
		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()

		event := integrationevent.WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		assert.NoError(t, dao.CreateWallet(context.Background(), event))

		var (
			retrievedWalletID string
			retrievedHolderID string
			balance           int
		)
		postgreSQLDB.QueryRow("SELECT wallet_id, holder_id, balance_in_cents FROM wallet_projections WHERE wallet_id = $1", walletID).Scan(
			&retrievedWalletID,
			&retrievedHolderID,
			&balance,
		)

		assert.Equal(t, walletID, retrievedWalletID)
		assert.Equal(t, holderID, retrievedHolderID)
		assert.Equal(t, 0, balance)
	})

	t.Run("It should return error when wallet_id already exists", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletPostgreSQLProjectionDAO(postgreSQLDB)
		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()

		event := integrationevent.WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.NoError(t, dao.CreateWallet(context.Background(), event))

		assert.Error(t, dao.CreateWallet(context.Background(), event))
	})

	t.Run("It should initialize wallet balance as zero", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletPostgreSQLProjectionDAO(postgreSQLDB)
		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()

		event := integrationevent.WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		assert.NoError(t, dao.CreateWallet(context.Background(), event))

		assert.Equal(t, 0, getTestWalletBalance(t, walletID))
	})
}

func TestWalletPostgreSQLProjectionDAO_ApplyFundsTransferred(t *testing.T) {
	t.Parallel()

	t.Run("It should successfully deduct balance when funds are transferred", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletPostgreSQLProjectionDAO(postgreSQLDB)
		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()

		createTestWallet(t, walletID, holderID, dao)

		postgreSQLDB.Exec("UPDATE wallet_projections SET balance_in_cents = $1 WHERE wallet_id = $2", 5000, walletID)

		toWalletID := uuid.NewUUID()
		createTestWallet(t, toWalletID, uuid.NewUUID(), dao)

		event := integrationevent.FundsTransferredEvent{
			TransferID:    uuid.NewUUID(),
			FromWalletID:  walletID,
			ToWalletID:    toWalletID,
			AmountInCents: 1000,
			Timestamp:     time.Now(),
		}
		assert.NoError(t, dao.ApplyFundsTransferred(context.Background(), event))

		assert.Equal(t, 4000, getTestWalletBalance(t, walletID))
	})

	t.Run("It should result in negative balance if deducted amount exceeds current balance", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletPostgreSQLProjectionDAO(postgreSQLDB)
		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()

		createTestWallet(t, walletID, holderID, dao)

		postgreSQLDB.Exec("UPDATE wallet_projections SET balance_in_cents = $1 WHERE wallet_id = $2", 500, walletID)

		toWalletID := uuid.NewUUID()
		createTestWallet(t, toWalletID, uuid.NewUUID(), dao)

		event := integrationevent.FundsTransferredEvent{
			TransferID:    uuid.NewUUID(),
			FromWalletID:  walletID,
			ToWalletID:    toWalletID,
			AmountInCents: 1000,
			Timestamp:     time.Now(),
		}
		assert.NoError(t, dao.ApplyFundsTransferred(context.Background(), event))

		assert.Equal(t, -500, getTestWalletBalance(t, walletID))
	})

	t.Run("It should update the updated_at timestamp when funds are transferred", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletPostgreSQLProjectionDAO(postgreSQLDB)
		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()

		createTestWallet(t, walletID, holderID, dao)

		postgreSQLDB.Exec("UPDATE wallet_projections SET balance_in_cents = $1 WHERE wallet_id = $2", 1000, walletID)

		oldTime := time.Now().Add(-1 * time.Hour)
		newTime := time.Now()

		postgreSQLDB.Exec("UPDATE wallet_projections SET updated_at = $1 WHERE wallet_id = $2", oldTime, walletID)

		toWalletID := uuid.NewUUID()
		createTestWallet(t, toWalletID, uuid.NewUUID(), dao)

		event := integrationevent.FundsTransferredEvent{
			TransferID:    uuid.NewUUID(),
			FromWalletID:  walletID,
			ToWalletID:    toWalletID,
			AmountInCents: 200,
			Timestamp:     newTime,
		}
		assert.NoError(t, dao.ApplyFundsTransferred(context.Background(), event))

		var updatedAt time.Time
		postgreSQLDB.QueryRow("SELECT updated_at FROM wallet_projections WHERE wallet_id = $1", walletID).Scan(&updatedAt)

		assert.True(t, updatedAt.After(oldTime))
	})

	t.Run("It should insert an outgoing transfer projection when funds are transferred", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletPostgreSQLProjectionDAO(postgreSQLDB)
		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()
		toWalletID := uuid.NewUUID()
		transferID := uuid.NewUUID()

		createTestWallet(t, walletID, holderID, dao)
		createTestWallet(t, toWalletID, uuid.NewUUID(), dao)

		postgreSQLDB.Exec("UPDATE wallet_projections SET balance_in_cents = $1 WHERE wallet_id = $2", 5000, walletID)

		event := integrationevent.FundsTransferredEvent{
			TransferID:    transferID,
			FromWalletID:  walletID,
			ToWalletID:    toWalletID,
			AmountInCents: 1000,
			Timestamp:     time.Now(),
		}
		assert.NoError(t, dao.ApplyFundsTransferred(context.Background(), event))

		row, found := getTestTransferProjection(t, walletID, transferID)
		assert.True(t, found)
		assert.Equal(t, 1000, row.AmountInCents)
		assert.Equal(t, "outgoing", row.Direction)
		assert.Equal(t, toWalletID, row.CounterpartWalletID)
	})
}

func TestWalletPostgreSQLProjectionDAO_ApplyFundsTransferReceived(t *testing.T) {
	t.Parallel()

	t.Run("It should successfully increase balance when funds are received", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletPostgreSQLProjectionDAO(postgreSQLDB)
		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()

		createTestWallet(t, walletID, holderID, dao)

		postgreSQLDB.Exec("UPDATE wallet_projections SET balance_in_cents = $1 WHERE wallet_id = $2", 1000, walletID)

		fromWalletID := uuid.NewUUID()
		createTestWallet(t, fromWalletID, uuid.NewUUID(), dao)

		event := integrationevent.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    uuid.NewUUID(),
			FromWalletID:  fromWalletID,
			AmountInCents: 500,
			Timestamp:     time.Now(),
		}
		assert.NoError(t, dao.ApplyFundsTransferReceived(context.Background(), event))

		assert.Equal(t, 1500, getTestWalletBalance(t, walletID))
	})

	t.Run("It should successfully add funds to zero balance wallet", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletPostgreSQLProjectionDAO(postgreSQLDB)
		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()

		createTestWallet(t, walletID, holderID, dao)

		fromWalletID := uuid.NewUUID()
		createTestWallet(t, fromWalletID, uuid.NewUUID(), dao)

		event := integrationevent.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    uuid.NewUUID(),
			FromWalletID:  fromWalletID,
			AmountInCents: 2000,
			Timestamp:     time.Now(),
		}
		assert.NoError(t, dao.ApplyFundsTransferReceived(context.Background(), event))

		assert.Equal(t, 2000, getTestWalletBalance(t, walletID))
	})

	t.Run("It should update the updated_at timestamp when funds are received", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletPostgreSQLProjectionDAO(postgreSQLDB)
		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()

		createTestWallet(t, walletID, holderID, dao)

		oldTime := time.Now().Add(-1 * time.Hour)
		newTime := time.Now()

		postgreSQLDB.Exec("UPDATE wallet_projections SET updated_at = $1 WHERE wallet_id = $2", oldTime, walletID)

		fromWalletID := uuid.NewUUID()
		createTestWallet(t, fromWalletID, uuid.NewUUID(), dao)

		event := integrationevent.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    uuid.NewUUID(),
			FromWalletID:  fromWalletID,
			AmountInCents: 750,
			Timestamp:     newTime,
		}
		assert.NoError(t, dao.ApplyFundsTransferReceived(context.Background(), event))

		var updatedAt time.Time
		postgreSQLDB.QueryRow("SELECT updated_at FROM wallet_projections WHERE wallet_id = $1", walletID).Scan(&updatedAt)

		assert.True(t, updatedAt.After(oldTime))
	})

	t.Run("It should successfully accumulate multiple fund transfers", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletPostgreSQLProjectionDAO(postgreSQLDB)
		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()

		createTestWallet(t, walletID, holderID, dao)

		fromWalletID1 := uuid.NewUUID()
		createTestWallet(t, fromWalletID1, uuid.NewUUID(), dao)

		event1 := integrationevent.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    uuid.NewUUID(),
			FromWalletID:  fromWalletID1,
			AmountInCents: 1000,
			Timestamp:     time.Now(),
		}
		assert.NoError(t, dao.ApplyFundsTransferReceived(context.Background(), event1))

		fromWalletID2 := uuid.NewUUID()
		createTestWallet(t, fromWalletID2, uuid.NewUUID(), dao)

		event2 := integrationevent.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    uuid.NewUUID(),
			FromWalletID:  fromWalletID2,
			AmountInCents: 500,
			Timestamp:     time.Now(),
		}
		assert.NoError(t, dao.ApplyFundsTransferReceived(context.Background(), event2))

		assert.Equal(t, 1500, getTestWalletBalance(t, walletID))
	})

	t.Run("It should insert an incoming transfer projection when funds are received", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletPostgreSQLProjectionDAO(postgreSQLDB)
		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()
		fromWalletID := uuid.NewUUID()
		transferID := uuid.NewUUID()

		createTestWallet(t, walletID, holderID, dao)
		createTestWallet(t, fromWalletID, uuid.NewUUID(), dao)

		event := integrationevent.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    transferID,
			FromWalletID:  fromWalletID,
			AmountInCents: 2000,
			Timestamp:     time.Now(),
		}
		assert.NoError(t, dao.ApplyFundsTransferReceived(context.Background(), event))

		row, found := getTestTransferProjection(t, walletID, transferID)
		assert.True(t, found)
		assert.Equal(t, 2000, row.AmountInCents)
		assert.Equal(t, "incoming", row.Direction)
		assert.Equal(t, fromWalletID, row.CounterpartWalletID)
	})
}
