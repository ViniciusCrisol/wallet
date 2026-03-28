package consumer

import (
	"encoding/json"
	"testing"
	"time"

	"wallet/wallet-service/internal/infrastructure/persistence"
	"wallet/wallet-service/pkg"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func createTestWallet(t *testing.T, dao *persistence.WalletMysqlDAO, walletID, holderID string) {
	event := pkg.WalletCreatedEvent{
		WalletID:  walletID,
		HolderID:  holderID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	assert.NoError(t, dao.CreateWallet(event))
}

func getWalletBalance(t *testing.T, walletID string) int {
	var balance int
	assert.NoError(t, testDB.QueryRow("SELECT balance_in_cents FROM wallet_projections WHERE wallet_id = ?", walletID).Scan(&balance))
	return balance
}

func TestProjectionConsumer_Handle(t *testing.T) {
	t.Run("It should process WalletCreatedEvent and persist wallet to database", func(t *testing.T) {
		dao := persistence.NewWalletMysqlDAO(testDB)
		consumer := NewProjectionConsumer(nil, dao)

		walletID := uuid.New().String()
		holderID := uuid.New().String()
		now := time.Now()

		event := pkg.WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		body, err := json.Marshal(event)
		assert.NoError(t, err)
		assert.NoError(t, consumer.handle(body, pkg.WalletCreatedEventName))

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

	t.Run("It should process FundsTransferredEvent and deduct balance from wallet", func(t *testing.T) {
		dao := persistence.NewWalletMysqlDAO(testDB)
		consumer := NewProjectionConsumer(nil, dao)

		walletID := uuid.New().String()
		holderID := uuid.New().String()
		createTestWallet(t, dao, walletID, holderID)

		testDB.Exec("UPDATE wallet_projections SET balance_in_cents = ? WHERE wallet_id = ?", 5000, walletID)

		toWalletID := uuid.New().String()
		transferID := uuid.New().String()
		now := time.Now()

		event := pkg.FundsTransferredEvent{
			TransferID:    transferID,
			ToWalletID:    toWalletID,
			FromWalletID:  walletID,
			AmountInCents: 1000,
			Timestamp:     now,
		}
		body, err := json.Marshal(event)
		assert.NoError(t, err)
		assert.NoError(t, consumer.handle(body, pkg.FundsTransferredEventName))

		assert.Equal(t, 4000, getWalletBalance(t, walletID))
	})

	t.Run("It should process FundsTransferReceivedEvent and add balance to wallet", func(t *testing.T) {
		dao := persistence.NewWalletMysqlDAO(testDB)
		consumer := NewProjectionConsumer(nil, dao)

		walletID := uuid.New().String()
		holderID := uuid.New().String()
		createTestWallet(t, dao, walletID, holderID)

		testDB.Exec("UPDATE wallet_projections SET balance_in_cents = ? WHERE wallet_id = ?", 1000, walletID)

		fromWalletID := uuid.New().String()
		transferID := uuid.New().String()
		now := time.Now()

		event := pkg.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    transferID,
			FromWalletID:  fromWalletID,
			AmountInCents: 2500,
			Timestamp:     now,
		}
		body, err := json.Marshal(event)
		assert.NoError(t, err)
		assert.NoError(t, consumer.handle(body, pkg.FundsTransferReceivedEventName))

		assert.Equal(t, 3500, getWalletBalance(t, walletID))
	})

	t.Run("It should return error when WalletCreatedEvent unmarshal fails", func(t *testing.T) {
		dao := persistence.NewWalletMysqlDAO(testDB)
		consumer := NewProjectionConsumer(nil, dao)

		invalidEventBody := []byte(`{invalid json}`)

		assert.Error(t, consumer.handle(invalidEventBody, pkg.WalletCreatedEventName))
	})

	t.Run("It should return error when FundsTransferredEvent unmarshal fails", func(t *testing.T) {
		dao := persistence.NewWalletMysqlDAO(testDB)
		consumer := NewProjectionConsumer(nil, dao)

		invalidEventBody := []byte(`{invalid json}`)

		assert.Error(t, consumer.handle(invalidEventBody, pkg.FundsTransferredEventName))
	})

	t.Run("It should return error when FundsTransferReceivedEvent unmarshal fails", func(t *testing.T) {
		dao := persistence.NewWalletMysqlDAO(testDB)
		consumer := NewProjectionConsumer(nil, dao)

		invalidEventBody := []byte(`{invalid json}`)

		assert.Error(t, consumer.handle(invalidEventBody, pkg.FundsTransferReceivedEventName))
	})

	t.Run("It should return nil error when unknown event type is received", func(t *testing.T) {
		dao := persistence.NewWalletMysqlDAO(testDB)
		consumer := NewProjectionConsumer(nil, dao)

		unknownEventBody := []byte(`{"some": "data"}`)

		assert.NoError(t, consumer.handle(unknownEventBody, "unknown:event_type"))
	})
}
