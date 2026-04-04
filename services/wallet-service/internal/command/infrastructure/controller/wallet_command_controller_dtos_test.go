package controller

import (
	"testing"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/pkg"
	valueObject "wallet/wallet-service/pkg/domain/value_object"

	"github.com/stretchr/testify/assert"
)

func TestCreateWalletDTO_CreateWalletCommand(t *testing.T) {
	t.Parallel()

	t.Run("It should return a valid command when both IDs are valid UUIDs", func(t *testing.T) {
		t.Parallel()

		walletID := valueObject.GenerateID().String()
		holderID := valueObject.GenerateID().String()
		dto := CreateWalletDTO{WalletID: walletID, HolderID: holderID, CreatedAt: "2025-01-15T10:30:00Z"}

		command, err := dto.CreateWalletCommand()

		assert.NoError(t, err)
		assert.Equal(t, walletID, command.WalletID.String())
		assert.Equal(t, holderID, command.HolderID.String())
		assert.False(t, command.Timestamp.IsZero())
		assert.Equal(t, 2025, command.Timestamp.Year())
	})

	t.Run("It should return an error when wallet_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		dto := CreateWalletDTO{WalletID: "not-a-uuid", HolderID: valueObject.GenerateID().String(), CreatedAt: "2025-01-15T10:30:00Z"}

		_, err := dto.CreateWalletCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidWalletID)
	})

	t.Run("It should return an error when holder_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		dto := CreateWalletDTO{WalletID: valueObject.GenerateID().String(), HolderID: "not-a-uuid", CreatedAt: "2025-01-15T10:30:00Z"}

		_, err := dto.CreateWalletCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidHolderID)
	})

	t.Run("It should return an error when wallet_id is empty", func(t *testing.T) {
		t.Parallel()

		dto := CreateWalletDTO{WalletID: "", HolderID: valueObject.GenerateID().String(), CreatedAt: "2025-01-15T10:30:00Z"}

		_, err := dto.CreateWalletCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidWalletID)
	})

	t.Run("It should return an error when holder_id is empty", func(t *testing.T) {
		t.Parallel()

		dto := CreateWalletDTO{WalletID: valueObject.GenerateID().String(), HolderID: "", CreatedAt: "2025-01-15T10:30:00Z"}

		_, err := dto.CreateWalletCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidHolderID)
	})

	t.Run("It should return an error when created_at is not a valid RFC3339 timestamp", func(t *testing.T) {
		t.Parallel()

		dto := CreateWalletDTO{WalletID: valueObject.GenerateID().String(), HolderID: valueObject.GenerateID().String(), CreatedAt: "not-a-date"}

		_, err := dto.CreateWalletCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidCreatedAt)
	})

	t.Run("It should return an error when created_at is empty", func(t *testing.T) {
		t.Parallel()

		dto := CreateWalletDTO{WalletID: valueObject.GenerateID().String(), HolderID: valueObject.GenerateID().String(), CreatedAt: ""}

		_, err := dto.CreateWalletCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidCreatedAt)
	})
}

func TestTransferFundsDTO_TransferFundsCommand(t *testing.T) {
	t.Parallel()

	t.Run("It should return a valid command when all fields are valid", func(t *testing.T) {
		t.Parallel()

		transferID := valueObject.GenerateID().String()
		toWalletID := valueObject.GenerateID().String()
		dto := TransferFundsDTO{AmountInCents: 500, TransferID: transferID, ToWalletID: toWalletID, Category: "food", TransferredAt: "2025-01-15T10:30:00Z"}

		command, err := dto.TransferFundsCommand()

		assert.NoError(t, err)
		assert.Equal(t, 500, command.Amount.Amount())
		assert.Equal(t, transferID, command.TransferID.String())
		assert.Equal(t, toWalletID, command.ToWalletID.String())
		assert.Equal(t, domain.CategoryFood, command.Category)
		assert.False(t, command.Timestamp.IsZero())
		assert.Equal(t, 2025, command.Timestamp.Year())
	})

	t.Run("It should return an error when transfer_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		dto := TransferFundsDTO{AmountInCents: 100, TransferID: "not-a-uuid", ToWalletID: valueObject.GenerateID().String(), TransferredAt: "2025-01-15T10:30:00Z"}

		_, err := dto.TransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidTransferID)
	})

	t.Run("It should default category to unclassified when empty", func(t *testing.T) {
		t.Parallel()

		dto := TransferFundsDTO{AmountInCents: 100, TransferID: valueObject.GenerateID().String(), ToWalletID: valueObject.GenerateID().String(), TransferredAt: "2025-01-15T10:30:00Z"}

		command, err := dto.TransferFundsCommand()

		assert.NoError(t, err)
		assert.Equal(t, domain.CategoryUnclassified, command.Category)
	})

	t.Run("It should return an error when category is invalid", func(t *testing.T) {
		t.Parallel()

		dto := TransferFundsDTO{AmountInCents: 100, TransferID: valueObject.GenerateID().String(), ToWalletID: valueObject.GenerateID().String(), Category: "invalid", TransferredAt: "2025-01-15T10:30:00Z"}

		_, err := dto.TransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidCategory)
	})

	t.Run("It should return an error when to_wallet_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		dto := TransferFundsDTO{AmountInCents: 100, TransferID: valueObject.GenerateID().String(), ToWalletID: "not-a-uuid", TransferredAt: "2025-01-15T10:30:00Z"}

		_, err := dto.TransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidToWalletID)
	})

	t.Run("It should return an error when amount_in_cents is zero", func(t *testing.T) {
		t.Parallel()

		dto := TransferFundsDTO{AmountInCents: 0, TransferID: valueObject.GenerateID().String(), ToWalletID: valueObject.GenerateID().String(), TransferredAt: "2025-01-15T10:30:00Z"}

		_, err := dto.TransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrNonPositiveAmount)
	})

	t.Run("It should return an error when amount_in_cents is negative", func(t *testing.T) {
		t.Parallel()

		dto := TransferFundsDTO{AmountInCents: -1, TransferID: valueObject.GenerateID().String(), ToWalletID: valueObject.GenerateID().String(), TransferredAt: "2025-01-15T10:30:00Z"}

		_, err := dto.TransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrNegativeAmount)
	})

	t.Run("It should return an error when transferred_at is not a valid RFC3339 timestamp", func(t *testing.T) {
		t.Parallel()

		dto := TransferFundsDTO{AmountInCents: 100, TransferID: valueObject.GenerateID().String(), ToWalletID: valueObject.GenerateID().String(), TransferredAt: "not-a-date"}

		_, err := dto.TransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidTransferredAt)
	})

	t.Run("It should return an error when transferred_at is empty", func(t *testing.T) {
		t.Parallel()

		dto := TransferFundsDTO{AmountInCents: 100, TransferID: valueObject.GenerateID().String(), ToWalletID: valueObject.GenerateID().String(), TransferredAt: ""}

		_, err := dto.TransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidTransferredAt)
	})
}

func TestMockTransferDTO_ReceiveFundsTransferCommand(t *testing.T) {
	t.Parallel()

	t.Run("It should return a valid command when all fields are valid", func(t *testing.T) {
		t.Parallel()

		transferID := valueObject.GenerateID().String()
		fromWalletID := valueObject.GenerateID().String()
		dto := MockTransferDTO{
			AmountInCents: 300,
			TransferID:    transferID,
			FromWalletID:  fromWalletID,
			TransferredAt: "2025-01-15T10:30:00Z",
		}

		command, err := dto.ReceiveFundsTransferCommand()

		assert.NoError(t, err)
		assert.Equal(t, 300, command.Amount.Amount())
		assert.Equal(t, transferID, command.TransferID.String())
		assert.Equal(t, fromWalletID, command.FromWalletID.String())
		assert.Equal(t, domain.CategoryUnclassified, command.Category)
		assert.False(t, command.Timestamp.IsZero())
		assert.Equal(t, 2025, command.Timestamp.Year())
	})

	t.Run("It should return an error when transfer_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		dto := MockTransferDTO{
			AmountInCents: 100,
			TransferID:    "not-a-uuid",
			FromWalletID:  valueObject.GenerateID().String(),
			TransferredAt: "2025-01-15T10:30:00Z",
		}

		_, err := dto.ReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidTransferID)
	})

	t.Run("It should default category to unclassified when empty", func(t *testing.T) {
		t.Parallel()

		dto := MockTransferDTO{
			AmountInCents: 100,
			TransferID:    valueObject.GenerateID().String(),
			FromWalletID:  valueObject.GenerateID().String(),
			TransferredAt: "2025-01-15T10:30:00Z",
		}

		command, err := dto.ReceiveFundsTransferCommand()

		assert.NoError(t, err)
		assert.Equal(t, domain.CategoryUnclassified, command.Category)
	})

	t.Run("It should return an error when category is invalid", func(t *testing.T) {
		t.Parallel()

		dto := MockTransferDTO{
			AmountInCents: 100,
			TransferID:    valueObject.GenerateID().String(),
			FromWalletID:  valueObject.GenerateID().String(),
			Category:      "invalid",
			TransferredAt: "2025-01-15T10:30:00Z",
		}

		_, err := dto.ReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidCategory)
	})

	t.Run("It should return an error when from_wallet_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		dto := MockTransferDTO{
			AmountInCents: 100,
			TransferID:    valueObject.GenerateID().String(),
			FromWalletID:  "not-a-uuid",
			TransferredAt: "2025-01-15T10:30:00Z",
		}

		_, err := dto.ReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidFromWalletID)
	})

	t.Run("It should return an error when amount_in_cents is zero", func(t *testing.T) {
		t.Parallel()

		dto := MockTransferDTO{
			AmountInCents: 0,
			TransferID:    valueObject.GenerateID().String(),
			FromWalletID:  valueObject.GenerateID().String(),
			TransferredAt: "2025-01-15T10:30:00Z",
		}

		_, err := dto.ReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrNonPositiveAmount)
	})

	t.Run("It should return an error when amount_in_cents is negative", func(t *testing.T) {
		t.Parallel()

		dto := MockTransferDTO{
			AmountInCents: -1,
			TransferID:    valueObject.GenerateID().String(),
			FromWalletID:  valueObject.GenerateID().String(),
			TransferredAt: "2025-01-15T10:30:00Z",
		}

		_, err := dto.ReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrNegativeAmount)
	})

	t.Run("It should return an error when transferred_at is not a valid RFC3339 timestamp", func(t *testing.T) {
		t.Parallel()

		dto := MockTransferDTO{
			AmountInCents: 100,
			TransferID:    valueObject.GenerateID().String(),
			FromWalletID:  valueObject.GenerateID().String(),
			TransferredAt: "not-a-date",
		}

		_, err := dto.ReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidTransferredAt)
	})

	t.Run("It should return an error when transferred_at is empty", func(t *testing.T) {
		t.Parallel()

		dto := MockTransferDTO{
			AmountInCents: 100,
			TransferID:    valueObject.GenerateID().String(),
			FromWalletID:  valueObject.GenerateID().String(),
			TransferredAt: "",
		}

		_, err := dto.ReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidTransferredAt)
	})
}
