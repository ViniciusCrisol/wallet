package consumer

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/internal/command/infrastructure/persistence"
	appErr "wallet/wallet-service/pkg/app_err"
	integrationEvent "wallet/wallet-service/pkg/platform/integration_event"
	"wallet/wallet-service/pkg/platform/uuid"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalletKurrentDBConsumer_Start(t *testing.T) {
	t.Parallel()

	t.Run("It should consume a published FundsTransferredEvent and apply ReceiveFundsTransfer to the target wallet", func(t *testing.T) {
		t.Parallel()

		group := "test-group-" + uuid.NewUUID()
		esHandler := persistence.NewWalletKurrentDBESHandler(client)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		client.CreatePersistentSubscriptionToAll(
			ctx,
			group,
			kurrentdb.PersistentAllSubscriptionOptions{},
		)

		targetWallet := createTestWalletWithBalance(t, esHandler, 0)

		go NewWalletKurrentDBConsumer(group, client, esHandler).Start(ctx)

		fromWalletID := uuid.NewUUID()
		transferID := uuid.NewUUID()
		now := time.Now()
		body, err := json.Marshal(integrationEvent.FundsTransferredEvent{
			TransferID:    transferID,
			ToWalletID:    targetWallet.ID().String(),
			FromWalletID:  fromWalletID,
			AmountInCents: 500,
			Category:      domain.CategoryFood,
			Timestamp:     now,
		})
		require.NoError(t, err)

		_, err = client.AppendToStream(
			ctx,
			"wallet-"+fromWalletID,
			kurrentdb.AppendToStreamOptions{},
			kurrentdb.EventData{
				Data:        body,
				EventID:     uuid.NewUUIDValue(),
				EventType:   integrationEvent.FundsTransferredEventName,
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
	t.Parallel()

	t.Run("It should process FundsTransferredEvent and credit the target wallet", func(t *testing.T) {
		t.Parallel()

		esHandler := persistence.NewWalletKurrentDBESHandler(client)
		targetWallet := createTestWalletWithBalance(t, esHandler, 1000)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		fromWalletID := uuid.NewUUID()
		transferID := uuid.NewUUID()
		now := time.Now()
		body, err := json.Marshal(integrationEvent.FundsTransferredEvent{
			TransferID:    transferID,
			ToWalletID:    targetWallet.ID().String(),
			FromWalletID:  fromWalletID,
			AmountInCents: 500,
			Category:      domain.CategoryFood,
			Timestamp:     now,
		})
		assert.NoError(t, err)

		assert.NoError(t, consumer.handle(context.Background(), body, integrationEvent.FundsTransferredEventName))

		wallet, found, err := esHandler.Find(context.Background(), targetWallet.ID())
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, 1500, wallet.Balance().Amount())
	})

	t.Run("It should return error when FundsTransferredEvent unmarshal fails", func(t *testing.T) {
		t.Parallel()

		esHandler := persistence.NewWalletKurrentDBESHandler(client)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		invalidEventBody := []byte(`{invalid json}`)

		assert.Error(t, consumer.handle(context.Background(), invalidEventBody, integrationEvent.FundsTransferredEventName))
	})

	t.Run("It should return nil error when unknown event type is received", func(t *testing.T) {
		t.Parallel()

		esHandler := persistence.NewWalletKurrentDBESHandler(client)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		unknownEventBody := []byte(`{"some": "data"}`)

		assert.NoError(t, consumer.handle(context.Background(), unknownEventBody, "unknown:event_type"))
	})

	t.Run("It should return error when to_wallet_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		esHandler := persistence.NewWalletKurrentDBESHandler(client)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		body, err := json.Marshal(integrationEvent.FundsTransferredEvent{
			TransferID:    uuid.NewUUID(),
			ToWalletID:    "not-a-uuid",
			FromWalletID:  uuid.NewUUID(),
			AmountInCents: 100,
			Timestamp:     time.Now(),
		})
		assert.NoError(t, err)

		assert.ErrorIs(t, consumer.handle(context.Background(), body, integrationEvent.FundsTransferredEventName), appErr.ErrInvalidUUID)
	})

	t.Run("It should return error when transfer_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		esHandler := persistence.NewWalletKurrentDBESHandler(client)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		body, err := json.Marshal(integrationEvent.FundsTransferredEvent{
			TransferID:    "not-a-uuid",
			ToWalletID:    uuid.NewUUID(),
			FromWalletID:  uuid.NewUUID(),
			AmountInCents: 100,
			Timestamp:     time.Now(),
		})
		assert.NoError(t, err)

		assert.ErrorIs(t, consumer.handle(context.Background(), body, integrationEvent.FundsTransferredEventName), appErr.ErrInvalidUUID)
	})

	t.Run("It should return error when from_wallet_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		esHandler := persistence.NewWalletKurrentDBESHandler(client)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		body, err := json.Marshal(integrationEvent.FundsTransferredEvent{
			TransferID:    uuid.NewUUID(),
			ToWalletID:    uuid.NewUUID(),
			FromWalletID:  "not-a-uuid",
			AmountInCents: 100,
			Timestamp:     time.Now(),
		})
		assert.NoError(t, err)

		assert.ErrorIs(t, consumer.handle(context.Background(), body, integrationEvent.FundsTransferredEventName), appErr.ErrInvalidUUID)
	})

	t.Run("It should return error when amount_in_cents is zero", func(t *testing.T) {
		t.Parallel()

		esHandler := persistence.NewWalletKurrentDBESHandler(client)
		targetWallet := createTestWalletWithBalance(t, esHandler, 1000)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		body, err := json.Marshal(integrationEvent.FundsTransferredEvent{
			TransferID:    uuid.NewUUID(),
			ToWalletID:    targetWallet.ID().String(),
			FromWalletID:  uuid.NewUUID(),
			AmountInCents: 0,
			Timestamp:     time.Now(),
		})
		assert.NoError(t, err)

		assert.NoError(t, consumer.handle(context.Background(), body, integrationEvent.FundsTransferredEventName))
	})

	t.Run("It should return error when amount_in_cents is negative", func(t *testing.T) {
		t.Parallel()

		esHandler := persistence.NewWalletKurrentDBESHandler(client)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		body, err := json.Marshal(integrationEvent.FundsTransferredEvent{
			TransferID:    uuid.NewUUID(),
			ToWalletID:    uuid.NewUUID(),
			FromWalletID:  uuid.NewUUID(),
			AmountInCents: -1,
			Timestamp:     time.Now(),
		})
		assert.NoError(t, err)

		assert.ErrorIs(t, consumer.handle(context.Background(), body, integrationEvent.FundsTransferredEventName), appErr.ErrNegativeAmount)
	})

	t.Run("It should return error when target wallet is not found", func(t *testing.T) {
		t.Parallel()

		esHandler := persistence.NewWalletKurrentDBESHandler(client)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		body, err := json.Marshal(integrationEvent.FundsTransferredEvent{
			TransferID:    uuid.NewUUID(),
			ToWalletID:    uuid.NewUUID(),
			FromWalletID:  uuid.NewUUID(),
			AmountInCents: 100,
			Timestamp:     time.Now(),
		})
		assert.NoError(t, err)

		assert.ErrorIs(t, consumer.handle(context.Background(), body, integrationEvent.FundsTransferredEventName), appErr.ErrWalletNotFound)
	})

	t.Run("It should return error when balance limit would be exceeded", func(t *testing.T) {
		t.Parallel()

		esHandler := persistence.NewWalletKurrentDBESHandler(client)
		targetWallet := createTestWalletWithBalance(t, esHandler, domain.MaxBalanceInCents)
		consumer := NewWalletKurrentDBConsumer("", nil, esHandler)

		body, err := json.Marshal(integrationEvent.FundsTransferredEvent{
			TransferID:    uuid.NewUUID(),
			ToWalletID:    targetWallet.ID().String(),
			FromWalletID:  uuid.NewUUID(),
			AmountInCents: 1,
			Timestamp:     time.Now(),
		})
		assert.NoError(t, err)

		assert.ErrorIs(t, consumer.handle(context.Background(), body, integrationEvent.FundsTransferredEventName), appErr.ErrBalanceLimitExceeded)
	})
}
