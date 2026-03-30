package consumer

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/internal/command/infrastructure/persistence"
	"wallet/wallet-service/pkg/apperr"
	"wallet/wallet-service/pkg/integrationevent"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/google/uuid"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestWalletWithBalance(t *testing.T, esHandler *persistence.WalletKurrentDBESHandler, balance int) domain.Wallet {
	t.Helper()

	wallet := domain.NewWallet(domain.CreateWalletCommand{
		WalletID:  valueobject.GenerateID(),
		HolderID:  valueobject.GenerateID(),
		Timestamp: time.Now(),
	})
	if balance > 0 {
		amount, err := valueobject.NewMoney(balance)
		require.NoError(t, err)
		require.NoError(t, wallet.ReceiveFundsTransfer(domain.ReceiveFundsTransferCommand{
			Amount:       amount,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		}))
	}
	require.NoError(t, esHandler.Save(context.Background(), wallet))
	return wallet
}

func TestWalletKurrentDBConsumer_Start(t *testing.T) {
	t.Run("It should consume a published FundsTransferredEvent and apply ReceiveFundsTransfer to the target wallet", func(t *testing.T) {
		group := "test-group-" + uuid.New().String()
		esHandler := persistence.NewWalletKurrentDBESHandler(kurrentDBClient)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		kurrentDBClient.CreatePersistentSubscriptionToAll(
			ctx,
			group,
			kurrentdb.PersistentAllSubscriptionOptions{},
		)

		targetWallet := createTestWalletWithBalance(t, esHandler, 0)

		go NewWalletKurrentDBConsumer(group, kurrentDBClient, esHandler).Start(ctx)

		fromWalletID := uuid.New().String()
		transferID := uuid.New().String()
		now := time.Now()
		body, err := json.Marshal(integrationevent.FundsTransferredEvent{
			TransferID:    transferID,
			ToWalletID:    targetWallet.ID().String(),
			FromWalletID:  fromWalletID,
			AmountInCents: 500,
			Timestamp:     now,
		})
		require.NoError(t, err)

		_, err = kurrentDBClient.AppendToStream(
			ctx,
			"wallet-"+fromWalletID,
			kurrentdb.AppendToStreamOptions{},
			kurrentdb.EventData{
				Data:        body,
				EventID:     uuid.New(),
				EventType:   integrationevent.FundsTransferredEventName,
				ContentType: kurrentdb.ContentTypeJson,
			},
		)
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			wallet, found, err := esHandler.Find(context.Background(), targetWallet.ID())
			if err != nil || !found {
				return false
			}
			return wallet.Balance().Amount() == 500
		}, time.Minute, time.Second)
	})
}

func TestWalletKurrentDBConsumer_Handle(t *testing.T) {
	t.Run("It should process FundsTransferredEvent and credit the target wallet", func(t *testing.T) {
		esHandler := persistence.NewWalletKurrentDBESHandler(kurrentDBClient)
		targetWallet := createTestWalletWithBalance(t, esHandler, 1000)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		fromWalletID := uuid.New().String()
		transferID := uuid.New().String()
		now := time.Now()
		body, err := json.Marshal(integrationevent.FundsTransferredEvent{
			TransferID:    transferID,
			ToWalletID:    targetWallet.ID().String(),
			FromWalletID:  fromWalletID,
			AmountInCents: 500,
			Timestamp:     now,
		})
		assert.NoError(t, err)

		assert.NoError(t, consumer.handle(context.Background(), body, integrationevent.FundsTransferredEventName))

		wallet, found, err := esHandler.Find(context.Background(), targetWallet.ID())
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, 1500, wallet.Balance().Amount())
	})

	t.Run("It should return error when FundsTransferredEvent unmarshal fails", func(t *testing.T) {
		esHandler := persistence.NewWalletKurrentDBESHandler(kurrentDBClient)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		invalidEventBody := []byte(`{invalid json}`)

		assert.Error(t, consumer.handle(context.Background(), invalidEventBody, integrationevent.FundsTransferredEventName))
	})

	t.Run("It should return nil error when unknown event type is received", func(t *testing.T) {
		esHandler := persistence.NewWalletKurrentDBESHandler(kurrentDBClient)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		unknownEventBody := []byte(`{"some": "data"}`)

		assert.NoError(t, consumer.handle(context.Background(), unknownEventBody, "unknown:event_type"))
	})

	t.Run("It should return error when to_wallet_id is not a valid UUID", func(t *testing.T) {
		esHandler := persistence.NewWalletKurrentDBESHandler(kurrentDBClient)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		body, err := json.Marshal(integrationevent.FundsTransferredEvent{
			TransferID:    uuid.New().String(),
			ToWalletID:    "not-a-uuid",
			FromWalletID:  uuid.New().String(),
			AmountInCents: 100,
			Timestamp:     time.Now(),
		})
		assert.NoError(t, err)

		assert.ErrorIs(t, consumer.handle(context.Background(), body, integrationevent.FundsTransferredEventName), apperr.ErrInvalidUUID)
	})

	t.Run("It should return error when transfer_id is not a valid UUID", func(t *testing.T) {
		esHandler := persistence.NewWalletKurrentDBESHandler(kurrentDBClient)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		body, err := json.Marshal(integrationevent.FundsTransferredEvent{
			TransferID:    "not-a-uuid",
			ToWalletID:    uuid.New().String(),
			FromWalletID:  uuid.New().String(),
			AmountInCents: 100,
			Timestamp:     time.Now(),
		})
		assert.NoError(t, err)

		assert.ErrorIs(t, consumer.handle(context.Background(), body, integrationevent.FundsTransferredEventName), apperr.ErrInvalidUUID)
	})

	t.Run("It should return error when from_wallet_id is not a valid UUID", func(t *testing.T) {
		esHandler := persistence.NewWalletKurrentDBESHandler(kurrentDBClient)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		body, err := json.Marshal(integrationevent.FundsTransferredEvent{
			TransferID:    uuid.New().String(),
			ToWalletID:    uuid.New().String(),
			FromWalletID:  "not-a-uuid",
			AmountInCents: 100,
			Timestamp:     time.Now(),
		})
		assert.NoError(t, err)

		assert.ErrorIs(t, consumer.handle(context.Background(), body, integrationevent.FundsTransferredEventName), apperr.ErrInvalidUUID)
	})

	t.Run("It should return error when amount_in_cents is zero", func(t *testing.T) {
		esHandler := persistence.NewWalletKurrentDBESHandler(kurrentDBClient)
		targetWallet := createTestWalletWithBalance(t, esHandler, 1000)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		body, err := json.Marshal(integrationevent.FundsTransferredEvent{
			TransferID:    uuid.New().String(),
			ToWalletID:    targetWallet.ID().String(),
			FromWalletID:  uuid.New().String(),
			AmountInCents: 0,
			Timestamp:     time.Now(),
		})
		assert.NoError(t, err)

		assert.NoError(t, consumer.handle(context.Background(), body, integrationevent.FundsTransferredEventName))
	})

	t.Run("It should return error when amount_in_cents is negative", func(t *testing.T) {
		esHandler := persistence.NewWalletKurrentDBESHandler(kurrentDBClient)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		body, err := json.Marshal(integrationevent.FundsTransferredEvent{
			TransferID:    uuid.New().String(),
			ToWalletID:    uuid.New().String(),
			FromWalletID:  uuid.New().String(),
			AmountInCents: -1,
			Timestamp:     time.Now(),
		})
		assert.NoError(t, err)

		assert.ErrorIs(t, consumer.handle(context.Background(), body, integrationevent.FundsTransferredEventName), apperr.ErrNegativeAmount)
	})

	t.Run("It should return error when target wallet is not found", func(t *testing.T) {
		esHandler := persistence.NewWalletKurrentDBESHandler(kurrentDBClient)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		body, err := json.Marshal(integrationevent.FundsTransferredEvent{
			TransferID:    uuid.New().String(),
			ToWalletID:    uuid.New().String(),
			FromWalletID:  uuid.New().String(),
			AmountInCents: 100,
			Timestamp:     time.Now(),
		})
		assert.NoError(t, err)

		assert.ErrorIs(t, consumer.handle(context.Background(), body, integrationevent.FundsTransferredEventName), apperr.ErrWalletNotFound)
	})

	t.Run("It should return error when balance limit would be exceeded", func(t *testing.T) {
		esHandler := persistence.NewWalletKurrentDBESHandler(kurrentDBClient)
		targetWallet := createTestWalletWithBalance(t, esHandler, domain.MaxBalanceInCents)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		body, err := json.Marshal(integrationevent.FundsTransferredEvent{
			TransferID:    uuid.New().String(),
			ToWalletID:    targetWallet.ID().String(),
			FromWalletID:  uuid.New().String(),
			AmountInCents: 1,
			Timestamp:     time.Now(),
		})
		assert.NoError(t, err)

		assert.ErrorIs(t, consumer.handle(context.Background(), body, integrationevent.FundsTransferredEventName), apperr.ErrBalanceLimitExceeded)
	})
}
