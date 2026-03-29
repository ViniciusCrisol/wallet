package projectionbuilder

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"wallet/wallet-service/pkg"

	"github.com/google/uuid"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalletKurrentDBProjectionConsumer_Start(t *testing.T) {
	t.Run("It should consume a published WalletCreatedEvent and persist the projection", func(t *testing.T) {
		groupName := "test-group-" + uuid.New().String()
		t.Setenv("WALLET_PROJECTION_GROUP_NAME", groupName)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		kurrentDBClient.CreatePersistentSubscriptionToAll(
			ctx,
			groupName,
			kurrentdb.PersistentAllSubscriptionOptions{},
		)
		go NewWalletKurrentDBProjectionConsumer(
			kurrentDBClient,
			NewWalletMySQLProjectionDAO(db),
		).Start(ctx)

		now := time.Now()
		walletID := uuid.New().String()
		holderID := uuid.New().String()
		body, err := json.Marshal(pkg.WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: now,
			UpdatedAt: now,
		})
		require.NoError(t, err)

		_, err = kurrentDBClient.AppendToStream(
			ctx,
			"wallet-"+walletID,
			kurrentdb.AppendToStreamOptions{},
			kurrentdb.EventData{
				Data:        body,
				EventID:     uuid.New(),
				EventType:   pkg.WalletCreatedEventName,
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
		groupName := "test-group-" + uuid.New().String()
		t.Setenv("WALLET_PROJECTION_GROUP_NAME", groupName)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		kurrentDBClient.CreatePersistentSubscriptionToAll(
			ctx,
			groupName,
			kurrentdb.PersistentAllSubscriptionOptions{},
		)

		walletMySQLProjectionDAO := NewWalletMySQLProjectionDAO(db)

		walletID := uuid.New().String()
		holderID := uuid.New().String()
		createTestWallet(t, walletID, holderID, walletMySQLProjectionDAO)
		db.Exec("UPDATE wallet_projections SET balance_in_cents = ? WHERE wallet_id = ?", 5000, walletID)

		go NewWalletKurrentDBProjectionConsumer(
			kurrentDBClient,
			walletMySQLProjectionDAO,
		).Start(ctx)

		now := time.Now()
		transferID := uuid.New().String()
		toWalletID := uuid.New().String()
		body, err := json.Marshal(pkg.FundsTransferredEvent{
			TransferID:    transferID,
			ToWalletID:    toWalletID,
			FromWalletID:  walletID,
			AmountInCents: 1000,
			Timestamp:     now,
		})
		require.NoError(t, err)

		_, err = kurrentDBClient.AppendToStream(
			ctx,
			"wallet-"+walletID,
			kurrentdb.AppendToStreamOptions{},
			kurrentdb.EventData{
				Data:        body,
				EventID:     uuid.New(),
				EventType:   pkg.FundsTransferredEventName,
				ContentType: kurrentdb.ContentTypeJson,
			},
		)
		require.NoError(t, err)

		require.Eventually(t, func() bool { return getTestWalletBalance(t, walletID) == 4000 }, time.Minute, time.Second)
	})

	t.Run("It should consume a published FundsTransferReceivedEvent and add balance to projection", func(t *testing.T) {
		groupName := "test-group-" + uuid.New().String()
		t.Setenv("WALLET_PROJECTION_GROUP_NAME", groupName)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		kurrentDBClient.CreatePersistentSubscriptionToAll(
			ctx,
			groupName,
			kurrentdb.PersistentAllSubscriptionOptions{},
		)

		walletMySQLProjectionDAO := NewWalletMySQLProjectionDAO(db)

		walletID := uuid.New().String()
		holderID := uuid.New().String()
		createTestWallet(t, walletID, holderID, walletMySQLProjectionDAO)
		db.Exec("UPDATE wallet_projections SET balance_in_cents = ? WHERE wallet_id = ?", 1000, walletID)

		go NewWalletKurrentDBProjectionConsumer(
			kurrentDBClient,
			walletMySQLProjectionDAO,
		).Start(ctx)

		now := time.Now()
		transferID := uuid.New().String()
		fromWalletID := uuid.New().String()
		body, err := json.Marshal(pkg.FundsTransferReceivedEvent{
			WalletID:      walletID,
			TransferID:    transferID,
			FromWalletID:  fromWalletID,
			AmountInCents: 2500,
			Timestamp:     now,
		})
		require.NoError(t, err)

		_, err = kurrentDBClient.AppendToStream(
			ctx,
			"wallet-"+walletID,
			kurrentdb.AppendToStreamOptions{},
			kurrentdb.EventData{
				Data:        body,
				EventID:     uuid.New(),
				EventType:   pkg.FundsTransferReceivedEventName,
				ContentType: kurrentdb.ContentTypeJson,
			},
		)
		require.NoError(t, err)

		require.Eventually(t, func() bool { return getTestWalletBalance(t, walletID) == 3500 }, time.Minute, time.Second)
	})
}

func TestWalletKurrentDBProjectionConsumer_Handle(t *testing.T) {
	t.Run("It should process WalletCreatedEvent and persist wallet to database", func(t *testing.T) {
		walletMySQLProjectionDAO := NewWalletMySQLProjectionDAO(db)
		walletKurrentDBProjectionConsumer := NewWalletKurrentDBProjectionConsumer(nil, walletMySQLProjectionDAO)

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
		assert.NoError(t, walletKurrentDBProjectionConsumer.handle(body, pkg.WalletCreatedEventName))

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
		walletMySQLProjectionDAO := NewWalletMySQLProjectionDAO(db)
		walletKurrentDBProjectionConsumer := NewWalletKurrentDBProjectionConsumer(nil, walletMySQLProjectionDAO)

		walletID := uuid.New().String()
		holderID := uuid.New().String()
		createTestWallet(t, walletID, holderID, walletMySQLProjectionDAO)

		db.Exec("UPDATE wallet_projections SET balance_in_cents = ? WHERE wallet_id = ?", 5000, walletID)

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
		assert.NoError(t, walletKurrentDBProjectionConsumer.handle(body, pkg.FundsTransferredEventName))

		assert.Equal(t, 4000, getTestWalletBalance(t, walletID))
	})

	t.Run("It should process FundsTransferReceivedEvent and add balance to wallet", func(t *testing.T) {
		walletMySQLProjectionDAO := NewWalletMySQLProjectionDAO(db)
		walletKurrentDBProjectionConsumer := NewWalletKurrentDBProjectionConsumer(nil, walletMySQLProjectionDAO)

		walletID := uuid.New().String()
		holderID := uuid.New().String()
		createTestWallet(t, walletID, holderID, walletMySQLProjectionDAO)

		db.Exec("UPDATE wallet_projections SET balance_in_cents = ? WHERE wallet_id = ?", 1000, walletID)

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
		assert.NoError(t, walletKurrentDBProjectionConsumer.handle(body, pkg.FundsTransferReceivedEventName))

		assert.Equal(t, 3500, getTestWalletBalance(t, walletID))
	})

	t.Run("It should return error when WalletCreatedEvent unmarshal fails", func(t *testing.T) {
		walletMySQLProjectionDAO := NewWalletMySQLProjectionDAO(db)
		walletKurrentDBProjectionConsumer := NewWalletKurrentDBProjectionConsumer(nil, walletMySQLProjectionDAO)

		invalidEventBody := []byte(`{invalid json}`)

		assert.Error(t, walletKurrentDBProjectionConsumer.handle(invalidEventBody, pkg.WalletCreatedEventName))
	})

	t.Run("It should return error when FundsTransferredEvent unmarshal fails", func(t *testing.T) {
		walletMySQLProjectionDAO := NewWalletMySQLProjectionDAO(db)
		walletKurrentDBProjectionConsumer := NewWalletKurrentDBProjectionConsumer(nil, walletMySQLProjectionDAO)

		invalidEventBody := []byte(`{invalid json}`)

		assert.Error(t, walletKurrentDBProjectionConsumer.handle(invalidEventBody, pkg.FundsTransferredEventName))
	})

	t.Run("It should return error when FundsTransferReceivedEvent unmarshal fails", func(t *testing.T) {
		walletMySQLProjectionDAO := NewWalletMySQLProjectionDAO(db)
		walletKurrentDBProjectionConsumer := NewWalletKurrentDBProjectionConsumer(nil, walletMySQLProjectionDAO)

		invalidEventBody := []byte(`{invalid json}`)

		assert.Error(t, walletKurrentDBProjectionConsumer.handle(invalidEventBody, pkg.FundsTransferReceivedEventName))
	})

	t.Run("It should return nil error when unknown event type is received", func(t *testing.T) {
		walletMySQLProjectionDAO := NewWalletMySQLProjectionDAO(db)
		walletKurrentDBProjectionConsumer := NewWalletKurrentDBProjectionConsumer(nil, walletMySQLProjectionDAO)

		unknownEventBody := []byte(`{"some": "data"}`)

		assert.NoError(t, walletKurrentDBProjectionConsumer.handle(unknownEventBody, "unknown:event_type"))
	})
}
