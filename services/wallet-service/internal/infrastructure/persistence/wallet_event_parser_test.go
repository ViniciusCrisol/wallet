package persistence

import (
	"testing"
	"time"

	"wallet/wallet-service/internal/domain"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/stretchr/testify/assert"
)

func TestParseWalletEvent(t *testing.T) {
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

		result, err := ParseWalletEvent(event)

		assert.NoError(t, err)
		assert.Equal(t, pkg.WalletCreatedEvent{
			WalletID:  walletID.ToString(),
			HolderID:  holderID.ToString(),
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

		result, err := ParseWalletEvent(event)

		assert.NoError(t, err)
		assert.Equal(t, pkg.FundsTransferredEvent{
			TransferID:    transferID.ToString(),
			ToWalletID:    walletID.ToString(),
			FromWalletID:  fromWalletID.ToString(),
			AmountInCents: amount.GetAmount(),
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

		result, err := ParseWalletEvent(event)

		assert.NoError(t, err)
		assert.Equal(t, pkg.FundsTransferReceivedEvent{
			WalletID:      walletID.ToString(),
			TransferID:    transferID.ToString(),
			FromWalletID:  fromWalletID.ToString(),
			AmountInCents: amount.GetAmount(),
			Timestamp:     now,
		}, result.Body)
		assert.Equal(t, pkg.FundsTransferReceivedEventName, result.Name)
	})

	t.Run("It should return an error when event type is unknown", func(t *testing.T) {
		type unknownEvent struct{}
		_, err := ParseWalletEvent(unknownEvent{})
		assert.ErrorIs(t, err, pkg.ErrUnknownEventType)
	})
}
