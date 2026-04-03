package projector

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	mySQLProjectorDAO "wallet/wallet-service/internal/projector/mysql_projector_dao"
	integrationEvent "wallet/wallet-service/pkg/platform/integration_event"
	"wallet/wallet-service/pkg/platform/uuid"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalletKurrentDBProjectorConsumer_Start(t *testing.T) {
	t.Parallel()

	t.Run("It should consume a published WalletCreatedEvent and persist the projection", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()
		body, err := json.Marshal(integrationEvent.WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: now,
			UpdatedAt: now,
		})
		require.NoError(t, err)

		_, err = client.AppendToStream(
			context.Background(),
			"wallet-"+walletID,
			kurrentdb.AppendToStreamOptions{},
			kurrentdb.EventData{
				Data:        body,
				EventID:     uuid.NewUUIDValue(),
				EventType:   integrationEvent.WalletCreatedEventName,
				ContentType: kurrentdb.ContentTypeJson,
			},
		)
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			var count int
			db.QueryRow("SELECT COUNT(*) FROM wallet_projections WHERE wallet_id = ?", walletID).Scan(&count)
			return count == 1
		}, time.Minute, time.Second)

		var (
			retrievedWalletID string
			retrievedHolderID string
			balance           int
		)
		db.QueryRow("SELECT wallet_id, holder_id, balance_in_cents FROM wallet_projections WHERE wallet_id = ?", walletID).Scan(
			&retrievedWalletID,
			&retrievedHolderID,
			&balance,
		)
		assert.Equal(t, walletID, retrievedWalletID)
		assert.Equal(t, holderID, retrievedHolderID)
		assert.Equal(t, 0, balance)
	})

	t.Run("It should consume a published FundsTransferredEvent and deduct balance from projection", func(t *testing.T) {
		t.Parallel()

		walletMySQLProjectionDAO := mySQLProjectorDAO.NewWalletMySQLProjectionDAO(db)

		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()
		createTestWallet(t, walletID, holderID, walletMySQLProjectionDAO)
		db.Exec("UPDATE wallet_projections SET balance_in_cents = ? WHERE wallet_id = ?", 5000, walletID)

		now := time.Now()
		transferID := uuid.NewUUID()
		toWalletID := uuid.NewUUID()
		createTestWallet(t, toWalletID, uuid.NewUUID(), walletMySQLProjectionDAO)
		body, err := json.Marshal(integrationEvent.FundsTransferredEvent{
			TransferID:    transferID,
			ToWalletID:    toWalletID,
			FromWalletID:  walletID,
			AmountInCents: 1000,
			Timestamp:     now,
		})
		require.NoError(t, err)

		_, err = client.AppendToStream(
			context.Background(),
			"wallet-"+walletID,
			kurrentdb.AppendToStreamOptions{},
			kurrentdb.EventData{
				Data:        body,
				EventID:     uuid.NewUUIDValue(),
				EventType:   integrationEvent.FundsTransferredEventName,
				ContentType: kurrentdb.ContentTypeJson,
			},
		)
		require.NoError(t, err)

		require.Eventually(t, func() bool { return getTestWalletBalance(t, walletID) == 4000 }, time.Minute, time.Second)

		require.Eventually(t, func() bool {
			_, found := getTestTransferProjection(t, walletID, transferID)
			return found
		}, time.Minute, time.Second)
		row, _ := getTestTransferProjection(t, walletID, transferID)
		assert.Equal(t, 1000, row.AmountInCents)
		assert.Equal(t, "outgoing", row.Direction)
		assert.Equal(t, toWalletID, row.CounterpartWalletID)
	})

	t.Run("It should consume a published FundsTransferReceivedEvent and add balance to projection", func(t *testing.T) {
		t.Parallel()

		walletMySQLProjectionDAO := mySQLProjectorDAO.NewWalletMySQLProjectionDAO(db)

		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()
		createTestWallet(t, walletID, holderID, walletMySQLProjectionDAO)
		db.Exec("UPDATE wallet_projections SET balance_in_cents = ? WHERE wallet_id = ?", 1000, walletID)

		now := time.Now()
		transferID := uuid.NewUUID()
		fromWalletID := uuid.NewUUID()
		createTestWallet(t, fromWalletID, uuid.NewUUID(), walletMySQLProjectionDAO)
		body, err := json.Marshal(integrationEvent.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    transferID,
			FromWalletID:  fromWalletID,
			AmountInCents: 2500,
			Timestamp:     now,
		})
		require.NoError(t, err)

		_, err = client.AppendToStream(
			context.Background(),
			"wallet-"+walletID,
			kurrentdb.AppendToStreamOptions{},
			kurrentdb.EventData{
				Data:        body,
				EventID:     uuid.NewUUIDValue(),
				EventType:   integrationEvent.FundsTransferReceivedEventName,
				ContentType: kurrentdb.ContentTypeJson,
			},
		)
		require.NoError(t, err)

		require.Eventually(t, func() bool { return getTestWalletBalance(t, walletID) == 3500 }, time.Minute, time.Second)

		require.Eventually(t, func() bool {
			_, found := getTestTransferProjection(t, walletID, transferID)
			return found
		}, time.Minute, time.Second)
		row, _ := getTestTransferProjection(t, walletID, transferID)
		assert.Equal(t, 2500, row.AmountInCents)
		assert.Equal(t, "incoming", row.Direction)
		assert.Equal(t, fromWalletID, row.CounterpartWalletID)
	})
}

func TestWalletKurrentDBProjectorConsumer_Handle(t *testing.T) {
	t.Parallel()

	t.Run("It should process WalletCreatedEvent and persist wallet to database", func(t *testing.T) {
		t.Parallel()

		walletMySQLProjectionDAO := mySQLProjectorDAO.NewWalletMySQLProjectionDAO(db)
		WalletKurrentDBProjectorConsumer := NewWalletKurrentDBProjectorConsumer("", nil, walletMySQLProjectionDAO)

		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()
		now := time.Now()

		event := integrationEvent.WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		body, err := json.Marshal(event)
		assert.NoError(t, err)
		assert.NoError(t, WalletKurrentDBProjectorConsumer.handle(context.Background(), body, integrationEvent.WalletCreatedEventName))

		var (
			retrievedWalletID string
			retrievedHolderID string
			balance           int
		)
		db.QueryRow("SELECT wallet_id, holder_id, balance_in_cents FROM wallet_projections WHERE wallet_id = ?", walletID).Scan(
			&retrievedWalletID,
			&retrievedHolderID,
			&balance,
		)

		assert.Equal(t, walletID, retrievedWalletID)
		assert.Equal(t, holderID, retrievedHolderID)
		assert.Equal(t, 0, balance)
	})

	t.Run("It should process FundsTransferredEvent and deduct balance from wallet", func(t *testing.T) {
		t.Parallel()

		walletMySQLProjectionDAO := mySQLProjectorDAO.NewWalletMySQLProjectionDAO(db)
		WalletKurrentDBProjectorConsumer := NewWalletKurrentDBProjectorConsumer("", nil, walletMySQLProjectionDAO)

		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()
		createTestWallet(t, walletID, holderID, walletMySQLProjectionDAO)

		db.Exec("UPDATE wallet_projections SET balance_in_cents = ? WHERE wallet_id = ?", 5000, walletID)

		now := time.Now()
		toWalletID := uuid.NewUUID()
		transferID := uuid.NewUUID()
		createTestWallet(t, toWalletID, uuid.NewUUID(), walletMySQLProjectionDAO)

		event := integrationEvent.FundsTransferredEvent{
			TransferID:    transferID,
			ToWalletID:    toWalletID,
			FromWalletID:  walletID,
			AmountInCents: 1000,
			Timestamp:     now,
		}
		body, err := json.Marshal(event)
		assert.NoError(t, err)
		assert.NoError(t, WalletKurrentDBProjectorConsumer.handle(context.Background(), body, integrationEvent.FundsTransferredEventName))

		assert.Equal(t, 4000, getTestWalletBalance(t, walletID))

		row, found := getTestTransferProjection(t, walletID, transferID)
		assert.True(t, found)
		assert.Equal(t, 1000, row.AmountInCents)
		assert.Equal(t, "outgoing", row.Direction)
		assert.Equal(t, toWalletID, row.CounterpartWalletID)
	})

	t.Run("It should process FundsTransferReceivedEvent and add balance to wallet", func(t *testing.T) {
		t.Parallel()

		walletMySQLProjectionDAO := mySQLProjectorDAO.NewWalletMySQLProjectionDAO(db)
		WalletKurrentDBProjectorConsumer := NewWalletKurrentDBProjectorConsumer("", nil, walletMySQLProjectionDAO)

		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()
		createTestWallet(t, walletID, holderID, walletMySQLProjectionDAO)

		db.Exec("UPDATE wallet_projections SET balance_in_cents = ? WHERE wallet_id = ?", 1000, walletID)

		now := time.Now()
		fromWalletID := uuid.NewUUID()
		transferID := uuid.NewUUID()
		createTestWallet(t, fromWalletID, uuid.NewUUID(), walletMySQLProjectionDAO)

		event := integrationEvent.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    transferID,
			FromWalletID:  fromWalletID,
			AmountInCents: 2500,
			Timestamp:     now,
		}
		body, err := json.Marshal(event)
		assert.NoError(t, err)
		assert.NoError(t, WalletKurrentDBProjectorConsumer.handle(context.Background(), body, integrationEvent.FundsTransferReceivedEventName))

		assert.Equal(t, 3500, getTestWalletBalance(t, walletID))

		row, found := getTestTransferProjection(t, walletID, transferID)
		assert.True(t, found)
		assert.Equal(t, 2500, row.AmountInCents)
		assert.Equal(t, "incoming", row.Direction)
		assert.Equal(t, fromWalletID, row.CounterpartWalletID)
	})

	t.Run("It should return error when WalletCreatedEvent unmarshal fails", func(t *testing.T) {
		t.Parallel()

		walletMySQLProjectionDAO := mySQLProjectorDAO.NewWalletMySQLProjectionDAO(db)
		WalletKurrentDBProjectorConsumer := NewWalletKurrentDBProjectorConsumer("", nil, walletMySQLProjectionDAO)

		invalidEventBody := []byte(`{invalid json}`)

		assert.Error(t, WalletKurrentDBProjectorConsumer.handle(context.Background(), invalidEventBody, integrationEvent.WalletCreatedEventName))
	})

	t.Run("It should return error when FundsTransferredEvent unmarshal fails", func(t *testing.T) {
		t.Parallel()

		walletMySQLProjectionDAO := mySQLProjectorDAO.NewWalletMySQLProjectionDAO(db)
		WalletKurrentDBProjectorConsumer := NewWalletKurrentDBProjectorConsumer("", nil, walletMySQLProjectionDAO)

		invalidEventBody := []byte(`{invalid json}`)

		assert.Error(t, WalletKurrentDBProjectorConsumer.handle(context.Background(), invalidEventBody, integrationEvent.FundsTransferredEventName))
	})

	t.Run("It should return error when FundsTransferReceivedEvent unmarshal fails", func(t *testing.T) {
		t.Parallel()

		walletMySQLProjectionDAO := mySQLProjectorDAO.NewWalletMySQLProjectionDAO(db)
		WalletKurrentDBProjectorConsumer := NewWalletKurrentDBProjectorConsumer("", nil, walletMySQLProjectionDAO)

		invalidEventBody := []byte(`{invalid json}`)

		assert.Error(t, WalletKurrentDBProjectorConsumer.handle(context.Background(), invalidEventBody, integrationEvent.FundsTransferReceivedEventName))
	})

	t.Run("It should return nil error when unknown event type is received", func(t *testing.T) {
		t.Parallel()

		walletMySQLProjectionDAO := mySQLProjectorDAO.NewWalletMySQLProjectionDAO(db)
		WalletKurrentDBProjectorConsumer := NewWalletKurrentDBProjectorConsumer("", nil, walletMySQLProjectionDAO)

		unknownEventBody := []byte(`{"some": "data"}`)

		assert.NoError(t, WalletKurrentDBProjectorConsumer.handle(context.Background(), unknownEventBody, "unknown:event_type"))
	})
}
