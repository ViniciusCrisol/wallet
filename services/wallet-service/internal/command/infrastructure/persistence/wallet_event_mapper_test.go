package persistence

import (
	"encoding/json"
	"testing"
	"time"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/pkg"
	valueObject "wallet/wallet-service/pkg/domain/value_object"
	integrationEvent "wallet/wallet-service/pkg/platform/integration_event"

	"github.com/stretchr/testify/assert"
)

func TestWalletDomainToIntegrationEvent(t *testing.T) {
	t.Parallel()

	now := time.Now()
	walletID := valueObject.GenerateID()
	holderID := valueObject.GenerateID()
	transferID := valueObject.GenerateID()
	fromWalletID := valueObject.GenerateID()
	amount, _ := valueObject.NewMoney(500)

	t.Run("It should return a ParsedEvent with WalletCreatedEvent name and body when event is WalletCreatedEvent", func(t *testing.T) {
		t.Parallel()

		event := domain.WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: now,
			UpdatedAt: now,
		}

		result, err := walletDomainToIntegrationEvent(event)

		assert.NoError(t, err)
		assert.Equal(t, integrationEvent.WalletCreatedEvent{
			WalletID:  walletID.String(),
			HolderID:  holderID.String(),
			CreatedAt: now,
			UpdatedAt: now,
		}, result.Body)
		assert.Equal(t, integrationEvent.WalletCreatedEventName, result.Name)
	})

	t.Run("It should return a ParsedEvent with FundsTransferredEvent name and body when event is FundsTransferredEvent", func(t *testing.T) {
		t.Parallel()

		event := domain.FundsTransferredEvent{
			TransferID:   transferID,
			ToWalletID:   walletID,
			FromWalletID: fromWalletID,
			Amount:       amount,
			Category:     domain.CategoryFood,
			Timestamp:    now,
		}

		result, err := walletDomainToIntegrationEvent(event)

		assert.NoError(t, err)
		assert.Equal(t, integrationEvent.FundsTransferredEvent{
			TransferID:    transferID.String(),
			ToWalletID:    walletID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: amount.Amount(),
			Category:      domain.CategoryFood,
			Timestamp:     now,
		}, result.Body)
		assert.Equal(t, integrationEvent.FundsTransferredEventName, result.Name)
	})

	t.Run("It should return a ParsedEvent with FundsTransferReceivedEvent name and body when event is FundsTransferReceivedEvent", func(t *testing.T) {
		t.Parallel()

		event := domain.FundsTransferReceivedEvent{
			WalletID:     walletID,
			TransferID:   transferID,
			FromWalletID: fromWalletID,
			Amount:       amount,
			Category:     domain.CategoryFuel,
			Timestamp:    now,
		}

		result, err := walletDomainToIntegrationEvent(event)

		assert.NoError(t, err)
		assert.Equal(t, integrationEvent.FundsTransferReceivedEvent{
			WalletID:      walletID.String(),
			TransferID:    transferID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: amount.Amount(),
			Category:      domain.CategoryFuel,
			Timestamp:     now,
		}, result.Body)
		assert.Equal(t, integrationEvent.FundsTransferReceivedEventName, result.Name)
	})

	t.Run("It should return an error when event type is unknown", func(t *testing.T) {
		t.Parallel()

		type unknownEvent struct{}
		_, err := walletDomainToIntegrationEvent(unknownEvent{})
		assert.ErrorIs(t, err, pkg.ErrUnknownEventType)
	})
}

func TestWalletIntegrationToDomainEvent(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Millisecond)
	walletID := valueObject.GenerateID()
	holderID := valueObject.GenerateID()
	transferID := valueObject.GenerateID()
	fromWalletID := valueObject.GenerateID()

	t.Run("It should return a WalletCreatedEvent when event name is WalletCreatedEventName and body is valid", func(t *testing.T) {
		t.Parallel()

		body, _ := json.Marshal(integrationEvent.WalletCreatedEvent{
			WalletID:  walletID.String(),
			HolderID:  holderID.String(),
			CreatedAt: now,
			UpdatedAt: now,
		})

		result, err := WalletIntegrationToDomainEvent(body, integrationEvent.WalletCreatedEventName)

		assert.NoError(t, err)
		event, ok := result.(domain.WalletCreatedEvent)
		assert.True(t, ok)
		assert.Equal(t, walletID.String(), event.WalletID.String())
		assert.Equal(t, holderID.String(), event.HolderID.String())
		assert.Equal(t, now, event.CreatedAt)
		assert.Equal(t, now, event.UpdatedAt)
	})

	t.Run("It should return an error when event name is WalletCreatedEventName and body is invalid JSON", func(t *testing.T) {
		t.Parallel()

		_, err := WalletIntegrationToDomainEvent([]byte("invalid"), integrationEvent.WalletCreatedEventName)
		assert.Error(t, err)
	})

	t.Run("It should return an error when event name is WalletCreatedEventName and WalletID is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		body, _ := json.Marshal(integrationEvent.WalletCreatedEvent{
			WalletID: "not-a-uuid",
			HolderID: holderID.String(),
		})
		_, err := WalletIntegrationToDomainEvent(body, integrationEvent.WalletCreatedEventName)
		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when event name is WalletCreatedEventName and HolderID is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		body, _ := json.Marshal(integrationEvent.WalletCreatedEvent{
			WalletID: walletID.String(),
			HolderID: "not-a-uuid",
		})
		_, err := WalletIntegrationToDomainEvent(body, integrationEvent.WalletCreatedEventName)
		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return a FundsTransferredEvent when event name is FundsTransferredEventName and body is valid", func(t *testing.T) {
		t.Parallel()

		body, _ := json.Marshal(integrationEvent.FundsTransferredEvent{
			TransferID:    transferID.String(),
			ToWalletID:    walletID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: 500,
			Category:      domain.CategoryFood,
			Timestamp:     now,
		})

		result, err := WalletIntegrationToDomainEvent(body, integrationEvent.FundsTransferredEventName)

		assert.NoError(t, err)
		event, ok := result.(domain.FundsTransferredEvent)
		assert.True(t, ok)
		assert.Equal(t, transferID.String(), event.TransferID.String())
		assert.Equal(t, walletID.String(), event.ToWalletID.String())
		assert.Equal(t, fromWalletID.String(), event.FromWalletID.String())
		assert.Equal(t, 500, event.Amount.Amount())
		assert.Equal(t, domain.CategoryFood, event.Category)
		assert.Equal(t, now, event.Timestamp)
	})

	t.Run("It should return an error when event name is FundsTransferredEventName and body is invalid JSON", func(t *testing.T) {
		t.Parallel()

		_, err := WalletIntegrationToDomainEvent([]byte("invalid"), integrationEvent.FundsTransferredEventName)
		assert.Error(t, err)
	})

	t.Run("It should return an error when event name is FundsTransferredEventName and TransferID is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		body, _ := json.Marshal(integrationEvent.FundsTransferredEvent{
			TransferID:    "not-a-uuid",
			ToWalletID:    walletID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: 500,
		})
		_, err := WalletIntegrationToDomainEvent(body, integrationEvent.FundsTransferredEventName)
		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when event name is FundsTransferredEventName and amount is invalid", func(t *testing.T) {
		t.Parallel()

		body, _ := json.Marshal(integrationEvent.FundsTransferredEvent{
			TransferID:    transferID.String(),
			ToWalletID:    walletID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: -1,
		})
		_, err := WalletIntegrationToDomainEvent(body, integrationEvent.FundsTransferredEventName)
		assert.ErrorIs(t, err, pkg.ErrNegativeAmount)
	})

	t.Run("It should return a FundsTransferReceivedEvent when event name is FundsTransferReceivedEventName and body is valid", func(t *testing.T) {
		t.Parallel()

		body, _ := json.Marshal(integrationEvent.FundsTransferReceivedEvent{
			WalletID:      walletID.String(),
			TransferID:    transferID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: 300,
			Category:      domain.CategoryFuel,
			Timestamp:     now,
		})

		result, err := WalletIntegrationToDomainEvent(body, integrationEvent.FundsTransferReceivedEventName)

		assert.NoError(t, err)
		event, ok := result.(domain.FundsTransferReceivedEvent)
		assert.True(t, ok)
		assert.Equal(t, walletID.String(), event.WalletID.String())
		assert.Equal(t, transferID.String(), event.TransferID.String())
		assert.Equal(t, fromWalletID.String(), event.FromWalletID.String())
		assert.Equal(t, 300, event.Amount.Amount())
		assert.Equal(t, domain.CategoryFuel, event.Category)
		assert.Equal(t, now, event.Timestamp)
	})

	t.Run("It should return an error when event name is FundsTransferReceivedEventName and body is invalid JSON", func(t *testing.T) {
		t.Parallel()

		_, err := WalletIntegrationToDomainEvent([]byte("invalid"), integrationEvent.FundsTransferReceivedEventName)
		assert.Error(t, err)
	})

	t.Run("It should return an error when event name is FundsTransferReceivedEventName and WalletID is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		body, _ := json.Marshal(integrationEvent.FundsTransferReceivedEvent{
			WalletID:      "not-a-uuid",
			TransferID:    transferID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: 300,
		})
		_, err := WalletIntegrationToDomainEvent(body, integrationEvent.FundsTransferReceivedEventName)
		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when event name is FundsTransferReceivedEventName and amount is invalid", func(t *testing.T) {
		t.Parallel()

		body, _ := json.Marshal(integrationEvent.FundsTransferReceivedEvent{
			WalletID:      walletID.String(),
			TransferID:    transferID.String(),
			FromWalletID:  fromWalletID.String(),
			AmountInCents: -1,
		})
		_, err := WalletIntegrationToDomainEvent(body, integrationEvent.FundsTransferReceivedEventName)
		assert.ErrorIs(t, err, pkg.ErrNegativeAmount)
	})

	t.Run("It should return an error when event name is unknown", func(t *testing.T) {
		t.Parallel()

		_, err := WalletIntegrationToDomainEvent([]byte("{}"), "unknown:event")
		assert.ErrorIs(t, err, pkg.ErrUnknownEventType)
	})
}
