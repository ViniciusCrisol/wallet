package persistence

import (
	"database/sql"
	"testing"
	"time"

	"wallet/wallet-service/pkg"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func createTestWallet(t *testing.T, dao *WalletMysqlDAO, walletID, holderID string) {
	event := pkg.WalletCreatedEvent{
		WalletID:  walletID,
		HolderID:  holderID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	assert.NoError(t, dao.CreateWallet(event))
}

func getWalletBalance(t *testing.T, db *sql.DB, walletID string) int {
	var balance int
	assert.NoError(t, db.QueryRow("SELECT balance_in_cents FROM wallet_projections WHERE wallet_id = ?", walletID).Scan(&balance))
	return balance
}

func TestWalletMysqlDAO_CreateWallet(t *testing.T) {
	t.Parallel()

	t.Run("It should successfully create a wallet when valid event is provided", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletMysqlDAO(testDB)
		walletID := uuid.New().String()
		holderID := uuid.New().String()

		event := pkg.WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		assert.NoError(t, dao.CreateWallet(event))

		var (
			retrievedWalletID string
			retrievedHolderID string
			balance           int
		)
		testDB.QueryRow("SELECT wallet_id, holder_id, balance_in_cents FROM wallet_projections WHERE wallet_id = ?", walletID).Scan(
			&retrievedWalletID,
			&retrievedHolderID,
			&balance,
		)

		assert.Equal(t, walletID, retrievedWalletID)
		assert.Equal(t, holderID, retrievedHolderID)
		assert.Equal(t, 0, balance)
	})

	t.Run("It should return an error when wallet_id already exists", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletMysqlDAO(testDB)
		walletID := uuid.New().String()
		holderID := uuid.New().String()

		event := pkg.WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.NoError(t, dao.CreateWallet(event))

		assert.Error(t, dao.CreateWallet(event))
	})

	t.Run("It should initialize wallet balance as zero", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletMysqlDAO(testDB)
		walletID := uuid.New().String()
		holderID := uuid.New().String()

		event := pkg.WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		assert.NoError(t, dao.CreateWallet(event))

		assert.Equal(t, 0, getWalletBalance(t, testDB, walletID))
	})
}

func TestWalletMysqlDAO_ApplyFundsTransferred(t *testing.T) {
	t.Parallel()

	t.Run("It should successfully deduct balance when funds are transferred", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletMysqlDAO(testDB)
		walletID := uuid.New().String()
		holderID := uuid.New().String()

		createTestWallet(t, dao, walletID, holderID)

		testDB.Exec("UPDATE wallet_projections SET balance_in_cents = ? WHERE wallet_id = ?", 5000, walletID)

		event := pkg.FundsTransferredEvent{
			TransferID:    uuid.New().String(),
			FromWalletID:  walletID,
			ToWalletID:    uuid.New().String(),
			AmountInCents: 1000,
			Timestamp:     time.Now(),
		}
		assert.NoError(t, dao.ApplyFundsTransferred(event))

		assert.Equal(t, 4000, getWalletBalance(t, testDB, walletID))
	})

	t.Run("It should result in negative balance if deducted amount exceeds current balance", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletMysqlDAO(testDB)
		walletID := uuid.New().String()
		holderID := uuid.New().String()

		createTestWallet(t, dao, walletID, holderID)

		testDB.Exec("UPDATE wallet_projections SET balance_in_cents = ? WHERE wallet_id = ?", 500, walletID)

		event := pkg.FundsTransferredEvent{
			TransferID:    uuid.New().String(),
			FromWalletID:  walletID,
			ToWalletID:    uuid.New().String(),
			AmountInCents: 1000,
			Timestamp:     time.Now(),
		}
		assert.NoError(t, dao.ApplyFundsTransferred(event))

		assert.Equal(t, -500, getWalletBalance(t, testDB, walletID))
	})

	t.Run("It should update the updated_at timestamp when funds are transferred", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletMysqlDAO(testDB)
		walletID := uuid.New().String()
		holderID := uuid.New().String()

		createTestWallet(t, dao, walletID, holderID)

		testDB.Exec("UPDATE wallet_projections SET balance_in_cents = ? WHERE wallet_id = ?", 1000, walletID)

		oldTime := time.Now().Add(-1 * time.Hour)
		newTime := time.Now()

		testDB.Exec("UPDATE wallet_projections SET updated_at = ? WHERE wallet_id = ?", oldTime, walletID)

		event := pkg.FundsTransferredEvent{
			TransferID:    uuid.New().String(),
			FromWalletID:  walletID,
			ToWalletID:    uuid.New().String(),
			AmountInCents: 200,
			Timestamp:     newTime,
		}
		assert.NoError(t, dao.ApplyFundsTransferred(event))

		var updatedAt time.Time
		testDB.QueryRow("SELECT updated_at FROM wallet_projections WHERE wallet_id = ?", walletID).Scan(&updatedAt)

		assert.True(t, updatedAt.After(oldTime))
	})
}

func TestWalletMysqlDAO_ApplyFundsTransferReceived(t *testing.T) {
	t.Parallel()

	t.Run("It should successfully increase balance when funds are received", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletMysqlDAO(testDB)
		walletID := uuid.New().String()
		holderID := uuid.New().String()

		createTestWallet(t, dao, walletID, holderID)

		testDB.Exec("UPDATE wallet_projections SET balance_in_cents = ? WHERE wallet_id = ?", 1000, walletID)

		event := pkg.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    uuid.New().String(),
			FromWalletID:  uuid.New().String(),
			AmountInCents: 500,
			Timestamp:     time.Now(),
		}
		assert.NoError(t, dao.ApplyFundsTransferReceived(event))

		assert.Equal(t, 1500, getWalletBalance(t, testDB, walletID))
	})

	t.Run("It should successfully add funds to zero balance wallet", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletMysqlDAO(testDB)
		walletID := uuid.New().String()
		holderID := uuid.New().String()

		createTestWallet(t, dao, walletID, holderID)

		event := pkg.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    uuid.New().String(),
			FromWalletID:  uuid.New().String(),
			AmountInCents: 2000,
			Timestamp:     time.Now(),
		}
		assert.NoError(t, dao.ApplyFundsTransferReceived(event))

		assert.Equal(t, 2000, getWalletBalance(t, testDB, walletID))
	})

	t.Run("It should update the updated_at timestamp when funds are received", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletMysqlDAO(testDB)
		walletID := uuid.New().String()
		holderID := uuid.New().String()

		createTestWallet(t, dao, walletID, holderID)

		oldTime := time.Now().Add(-1 * time.Hour)
		newTime := time.Now()

		testDB.Exec("UPDATE wallet_projections SET updated_at = ? WHERE wallet_id = ?", oldTime, walletID)

		event := pkg.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    uuid.New().String(),
			FromWalletID:  uuid.New().String(),
			AmountInCents: 750,
			Timestamp:     newTime,
		}
		assert.NoError(t, dao.ApplyFundsTransferReceived(event))

		var updatedAt time.Time
		testDB.QueryRow("SELECT updated_at FROM wallet_projections WHERE wallet_id = ?", walletID).Scan(&updatedAt)

		assert.True(t, updatedAt.After(oldTime))
	})

	t.Run("It should successfully accumulate multiple fund transfers", func(t *testing.T) {
		t.Parallel()

		dao := NewWalletMysqlDAO(testDB)
		walletID := uuid.New().String()
		holderID := uuid.New().String()

		createTestWallet(t, dao, walletID, holderID)

		event1 := pkg.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    uuid.New().String(),
			FromWalletID:  uuid.New().String(),
			AmountInCents: 1000,
			Timestamp:     time.Now(),
		}
		assert.NoError(t, dao.ApplyFundsTransferReceived(event1))

		event2 := pkg.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    uuid.New().String(),
			FromWalletID:  uuid.New().String(),
			AmountInCents: 500,
			Timestamp:     time.Now(),
		}
		assert.NoError(t, dao.ApplyFundsTransferReceived(event2))

		assert.Equal(t, 1500, getWalletBalance(t, testDB, walletID))
	})
}
