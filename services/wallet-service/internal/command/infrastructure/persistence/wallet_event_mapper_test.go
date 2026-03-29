package persistence

import (
	"encoding/json"
	"testing"
	"time"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/stretchr/testify/assert"
)

func TestWalletDomainToIntegrationEvent(t *testing.T) {
	now := time.Now()
	walletID := valueobject.GenerateID()
	holderID := valueobject.GenerateID()
	transferID := valueobject.GenerateID()
	fromWalletID := valueobject.GenerateID()
	amount, _ := valueobject.NewMoney(500)

	t.Run("It should return a ParsedEvent with WalletCreatedEvent name and body when event is WalletCreatedEvent", func(t *testing.T) {
		event := domain.WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: now,
			UpdatedAt: now,
		}

		result, err := WalletDomainToIntegrationEvent(event)

		assert.NoError(t, err)
		assert.Equal(t, pkg.WalletCreatedEvent{
			WalletID:  walletID.String(),
			HolderID:  holderID.String(),
			CreatedAt: now,
			UpdatedAt: now,
		}, result.Body)
		assert.Equal(t, pkg.WalletCreatedEventName, result.Name)
	})

	t.Run("It should return a ParsedEvent with FundsTransferredEvent name and body when event is FundsTransferredEvent", func(t *testing.T) {
		event := domain.FundsTransferredEvent{
			TransferID:   transferID,
			ToWalletID:   walletID,
			FromWalletID: fromWalletID,
			Amount:       amount,
			Timestamp:    now,
		}

		result, err := WalletDomainToIntegrationEvent(event)

		assert.NoError(t, err)
		assert.Equal(t, pkg.FundsTransferredEvent{
			TransferID:    transferID.String(),
			ToWalletID:    walletID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: amount.Amount(),
			Timestamp:     now,
		}, result.Body)
		assert.Equal(t, pkg.FundsTransferredEventName, result.Name)
	})

	t.Run("It should return a ParsedEvent with FundsTransferReceivedEvent name and body when event is FundsTransferReceivedEvent", func(t *testing.T) {
		event := domain.FundsTransferReceivedEvent{
			WalletID:     walletID,
			TransferID:   transferID,
			FromWalletID: fromWalletID,
			Amount:       amount,
			Timestamp:    now,
		}

		result, err := WalletDomainToIntegrationEvent(event)

		assert.NoError(t, err)
		assert.Equal(t, pkg.FundsTransferReceivedEvent{
			WalletID:      walletID.String(),
			TransferID:    transferID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: amount.Amount(),
			Timestamp:     now,
		}, result.Body)
		assert.Equal(t, pkg.FundsTransferReceivedEventName, result.Name)
	})

	t.Run("It should return an error when event type is unknown", func(t *testing.T) {
		type unknownEvent struct{}
		_, err := WalletDomainToIntegrationEvent(unknownEvent{})
		assert.ErrorIs(t, err, pkg.ErrUnknownEventType)
	})
}

func TestWalletIntegrationToDomainEvent(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	walletID := valueobject.GenerateID()
	holderID := valueobject.GenerateID()
	transferID := valueobject.GenerateID()
	fromWalletID := valueobject.GenerateID()

	t.Run("It should return a WalletCreatedEvent when event name is WalletCreatedEventName and body is valid", func(t *testing.T) {
		body, _ := json.Marshal(pkg.WalletCreatedEvent{
			WalletID:  walletID.String(),
			HolderID:  holderID.String(),
			CreatedAt: now,
			UpdatedAt: now,
		})

		result, err := WalletIntegrationToDomainEvent(body, pkg.WalletCreatedEventName)

		assert.NoError(t, err)
		event, ok := result.(domain.WalletCreatedEvent)
		assert.True(t, ok)
		assert.Equal(t, walletID.String(), event.WalletID.String())
		assert.Equal(t, holderID.String(), event.HolderID.String())
		assert.Equal(t, now, event.CreatedAt)
		assert.Equal(t, now, event.UpdatedAt)
	})

	t.Run("It should return an error when event name is WalletCreatedEventName and body is invalid JSON", func(t *testing.T) {
		_, err := WalletIntegrationToDomainEvent([]byte("invalid"), pkg.WalletCreatedEventName)
		assert.Error(t, err)
	})

	t.Run("It should return an error when event name is WalletCreatedEventName and WalletID is not a valid UUID", func(t *testing.T) {
		body, _ := json.Marshal(pkg.WalletCreatedEvent{
			WalletID: "not-a-uuid",
			HolderID: holderID.String(),
		})
		_, err := WalletIntegrationToDomainEvent(body, pkg.WalletCreatedEventName)
		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when event name is WalletCreatedEventName and HolderID is not a valid UUID", func(t *testing.T) {
		body, _ := json.Marshal(pkg.WalletCreatedEvent{
			WalletID: walletID.String(),
			HolderID: "not-a-uuid",
		})
		_, err := WalletIntegrationToDomainEvent(body, pkg.WalletCreatedEventName)
		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return a FundsTransferredEvent when event name is FundsTransferredEventName and body is valid", func(t *testing.T) {
		body, _ := json.Marshal(pkg.FundsTransferredEvent{
			TransferID:    transferID.String(),
			ToWalletID:    walletID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: 500,
			Timestamp:     now,
		})

		result, err := WalletIntegrationToDomainEvent(body, pkg.FundsTransferredEventName)

		assert.NoError(t, err)
		event, ok := result.(domain.FundsTransferredEvent)
		assert.True(t, ok)
		assert.Equal(t, transferID.String(), event.TransferID.String())
		assert.Equal(t, walletID.String(), event.ToWalletID.String())
		assert.Equal(t, fromWalletID.String(), event.FromWalletID.String())
		assert.Equal(t, 500, event.Amount.Amount())
		assert.Equal(t, now, event.Timestamp)
	})

	t.Run("It should return an error when event name is FundsTransferredEventName and body is invalid JSON", func(t *testing.T) {
		_, err := WalletIntegrationToDomainEvent([]byte("invalid"), pkg.FundsTransferredEventName)
		assert.Error(t, err)
	})

	t.Run("It should return an error when event name is FundsTransferredEventName and TransferID is not a valid UUID", func(t *testing.T) {
		body, _ := json.Marshal(pkg.FundsTransferredEvent{
			TransferID:    "not-a-uuid",
			ToWalletID:    walletID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: 500,
		})
		_, err := WalletIntegrationToDomainEvent(body, pkg.FundsTransferredEventName)
		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when event name is FundsTransferredEventName and amount is invalid", func(t *testing.T) {
		body, _ := json.Marshal(pkg.FundsTransferredEvent{
			TransferID:    transferID.String(),
			ToWalletID:    walletID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: 0,
		})
		_, err := WalletIntegrationToDomainEvent(body, pkg.FundsTransferredEventName)
		assert.ErrorIs(t, err, pkg.ErrNegativeOrZeroAmount)
	})

	t.Run("It should return a FundsTransferReceivedEvent when event name is FundsTransferReceivedEventName and body is valid", func(t *testing.T) {
		body, _ := json.Marshal(pkg.FundsTransferReceivedEvent{
			WalletID:      walletID.String(),
			TransferID:    transferID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: 300,
			Timestamp:     now,
		})

		result, err := WalletIntegrationToDomainEvent(body, pkg.FundsTransferReceivedEventName)

		assert.NoError(t, err)
		event, ok := result.(domain.FundsTransferReceivedEvent)
		assert.True(t, ok)
		assert.Equal(t, walletID.String(), event.WalletID.String())
		assert.Equal(t, transferID.String(), event.TransferID.String())
		assert.Equal(t, fromWalletID.String(), event.FromWalletID.String())
		assert.Equal(t, 300, event.Amount.Amount())
		assert.Equal(t, now, event.Timestamp)
	})

	t.Run("It should return an error when event name is FundsTransferReceivedEventName and body is invalid JSON", func(t *testing.T) {
		_, err := WalletIntegrationToDomainEvent([]byte("invalid"), pkg.FundsTransferReceivedEventName)
		assert.Error(t, err)
	})

	t.Run("It should return an error when event name is FundsTransferReceivedEventName and WalletID is not a valid UUID", func(t *testing.T) {
		body, _ := json.Marshal(pkg.FundsTransferReceivedEvent{
			WalletID:      "not-a-uuid",
			TransferID:    transferID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: 300,
		})
		_, err := WalletIntegrationToDomainEvent(body, pkg.FundsTransferReceivedEventName)
		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when event name is FundsTransferReceivedEventName and amount is invalid", func(t *testing.T) {
		body, _ := json.Marshal(pkg.FundsTransferReceivedEvent{
			WalletID:      walletID.String(),
			TransferID:    transferID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: 0,
		})
		_, err := WalletIntegrationToDomainEvent(body, pkg.FundsTransferReceivedEventName)
		assert.ErrorIs(t, err, pkg.ErrNegativeOrZeroAmount)
	})

	t.Run("It should return an error when event name is unknown", func(t *testing.T) {
		_, err := WalletIntegrationToDomainEvent([]byte("{}"), "unknown:event")
		assert.ErrorIs(t, err, pkg.ErrUnknownEventType)
	})
}
