package controller

import (
	"testing"

	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/stretchr/testify/assert"
)

func TestCreateWalletDTO_CreateWalletCommand(t *testing.T) {
	t.Run("It should return a valid command when both IDs are valid UUIDs", func(t *testing.T) {
		walletID := valueobject.GenerateID().String()
		holderID := valueobject.GenerateID().String()
		dto := CreateWalletDTO{WalletID: walletID, HolderID: holderID}

		command, err := dto.CreateWalletCommand()

		assert.NoError(t, err)
		assert.Equal(t, walletID, command.WalletID.String())
		assert.Equal(t, holderID, command.HolderID.String())
		assert.False(t, command.Timestamp.IsZero())
	})

	t.Run("It should return an error when wallet_id is not a valid UUID", func(t *testing.T) {
		dto := CreateWalletDTO{WalletID: "not-a-uuid", HolderID: valueobject.GenerateID().String()}

		_, err := dto.CreateWalletCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when holder_id is not a valid UUID", func(t *testing.T) {
		dto := CreateWalletDTO{WalletID: valueobject.GenerateID().String(), HolderID: "not-a-uuid"}

		_, err := dto.CreateWalletCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when wallet_id is empty", func(t *testing.T) {
		dto := CreateWalletDTO{WalletID: "", HolderID: valueobject.GenerateID().String()}

		_, err := dto.CreateWalletCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when holder_id is empty", func(t *testing.T) {
		dto := CreateWalletDTO{WalletID: valueobject.GenerateID().String(), HolderID: ""}

		_, err := dto.CreateWalletCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})
}

func TestTransferFundsDTO_TransferFundsCommand(t *testing.T) {
	t.Run("It should return a valid command when all fields are valid", func(t *testing.T) {
		transferID := valueobject.GenerateID().String()
		toWalletID := valueobject.GenerateID().String()
		dto := TransferFundsDTO{AmountInCents: 500, TransferID: transferID, ToWalletID: toWalletID}

		command, err := dto.TransferFundsCommand()

		assert.NoError(t, err)
		assert.Equal(t, 500, command.Amount.Amount())
		assert.Equal(t, transferID, command.TransferID.String())
		assert.Equal(t, toWalletID, command.ToWalletID.String())
		assert.False(t, command.Timestamp.IsZero())
	})

	t.Run("It should return an error when transfer_id is not a valid UUID", func(t *testing.T) {
		dto := TransferFundsDTO{AmountInCents: 100, TransferID: "not-a-uuid", ToWalletID: valueobject.GenerateID().String()}

		_, err := dto.TransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when to_wallet_id is not a valid UUID", func(t *testing.T) {
		dto := TransferFundsDTO{AmountInCents: 100, TransferID: valueobject.GenerateID().String(), ToWalletID: "not-a-uuid"}

		_, err := dto.TransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when amount_in_cents is zero", func(t *testing.T) {
		dto := TransferFundsDTO{AmountInCents: 0, TransferID: valueobject.GenerateID().String(), ToWalletID: valueobject.GenerateID().String()}

		_, err := dto.TransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrNegativeOrZeroAmount)
	})

	t.Run("It should return an error when amount_in_cents is negative", func(t *testing.T) {
		dto := TransferFundsDTO{AmountInCents: -1, TransferID: valueobject.GenerateID().String(), ToWalletID: valueobject.GenerateID().String()}

		_, err := dto.TransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrNegativeOrZeroAmount)
	})
}

func TestMockTransferDTO_ReceiveFundsTransferCommand(t *testing.T) {
	t.Run("It should return a valid command when all fields are valid", func(t *testing.T) {
		transferID := valueobject.GenerateID().String()
		fromWalletID := valueobject.GenerateID().String()
		dto := MockTransferDTO{
			AmountInCents: 300,
			WalletID:      valueobject.GenerateID().String(),
			TransferID:    transferID,
			FromWalletID:  fromWalletID,
		}

		command, err := dto.ReceiveFundsTransferCommand()

		assert.NoError(t, err)
		assert.Equal(t, 300, command.Amount.Amount())
		assert.Equal(t, transferID, command.TransferID.String())
		assert.Equal(t, fromWalletID, command.FromWalletID.String())
		assert.False(t, command.Timestamp.IsZero())
	})

	t.Run("It should return an error when transfer_id is not a valid UUID", func(t *testing.T) {
		dto := MockTransferDTO{
			AmountInCents: 100,
			WalletID:      valueobject.GenerateID().String(),
			TransferID:    "not-a-uuid",
			FromWalletID:  valueobject.GenerateID().String(),
		}

		_, err := dto.ReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when from_wallet_id is not a valid UUID", func(t *testing.T) {
		dto := MockTransferDTO{
			AmountInCents: 100,
			WalletID:      valueobject.GenerateID().String(),
			TransferID:    valueobject.GenerateID().String(),
			FromWalletID:  "not-a-uuid",
		}

		_, err := dto.ReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when amount_in_cents is zero", func(t *testing.T) {
		dto := MockTransferDTO{
			AmountInCents: 0,
			WalletID:      valueobject.GenerateID().String(),
			TransferID:    valueobject.GenerateID().String(),
			FromWalletID:  valueobject.GenerateID().String(),
		}

		_, err := dto.ReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrNegativeOrZeroAmount)
	})

	t.Run("It should return an error when amount_in_cents is negative", func(t *testing.T) {
		dto := MockTransferDTO{
			AmountInCents: -1,
			WalletID:      valueobject.GenerateID().String(),
			TransferID:    valueobject.GenerateID().String(),
			FromWalletID:  valueobject.GenerateID().String(),
		}

		_, err := dto.ReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrNegativeOrZeroAmount)
	})
}
