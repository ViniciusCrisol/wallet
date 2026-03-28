package controller

import (
	"testing"

	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/stretchr/testify/assert"
)

func TestCreateWalletDTO_ToCreateWalletCommand(t *testing.T) {
	t.Run("It should return a valid command when both IDs are valid UUIDs", func(t *testing.T) {
		walletID := valueobject.GenerateID().ToString()
		holderID := valueobject.GenerateID().ToString()
		dto := CreateWalletDTO{WalletID: walletID, HolderID: holderID}

		command, err := dto.ToCreateWalletCommand()

		assert.NoError(t, err)
		assert.Equal(t, walletID, command.WalletID.ToString())
		assert.Equal(t, holderID, command.HolderID.ToString())
		assert.False(t, command.Timestamp.IsZero())
	})

	t.Run("It should return an error when wallet_id is not a valid UUID", func(t *testing.T) {
		dto := CreateWalletDTO{WalletID: "not-a-uuid", HolderID: valueobject.GenerateID().ToString()}

		_, err := dto.ToCreateWalletCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when holder_id is not a valid UUID", func(t *testing.T) {
		dto := CreateWalletDTO{WalletID: valueobject.GenerateID().ToString(), HolderID: "not-a-uuid"}

		_, err := dto.ToCreateWalletCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when wallet_id is empty", func(t *testing.T) {
		dto := CreateWalletDTO{WalletID: "", HolderID: valueobject.GenerateID().ToString()}

		_, err := dto.ToCreateWalletCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when holder_id is empty", func(t *testing.T) {
		dto := CreateWalletDTO{WalletID: valueobject.GenerateID().ToString(), HolderID: ""}

		_, err := dto.ToCreateWalletCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})
}

func TestTransferFundsDTO_ToTransferFundsCommand(t *testing.T) {
	t.Run("It should return a valid command when all fields are valid", func(t *testing.T) {
		transferID := valueobject.GenerateID().ToString()
		toWalletID := valueobject.GenerateID().ToString()
		dto := TransferFundsDTO{AmountInCents: 500, TransferID: transferID, ToWalletID: toWalletID}

		command, err := dto.ToTransferFundsCommand()

		assert.NoError(t, err)
		assert.Equal(t, 500, command.Amount.GetAmount())
		assert.Equal(t, transferID, command.TransferID.ToString())
		assert.Equal(t, toWalletID, command.ToWalletID.ToString())
		assert.False(t, command.Timestamp.IsZero())
	})

	t.Run("It should return an error when transfer_id is not a valid UUID", func(t *testing.T) {
		dto := TransferFundsDTO{AmountInCents: 100, TransferID: "not-a-uuid", ToWalletID: valueobject.GenerateID().ToString()}

		_, err := dto.ToTransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when to_wallet_id is not a valid UUID", func(t *testing.T) {
		dto := TransferFundsDTO{AmountInCents: 100, TransferID: valueobject.GenerateID().ToString(), ToWalletID: "not-a-uuid"}

		_, err := dto.ToTransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when amount_in_cents is zero", func(t *testing.T) {
		dto := TransferFundsDTO{AmountInCents: 0, TransferID: valueobject.GenerateID().ToString(), ToWalletID: valueobject.GenerateID().ToString()}

		_, err := dto.ToTransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrNegativeOrZeroAmount)
	})

	t.Run("It should return an error when amount_in_cents is negative", func(t *testing.T) {
		dto := TransferFundsDTO{AmountInCents: -1, TransferID: valueobject.GenerateID().ToString(), ToWalletID: valueobject.GenerateID().ToString()}

		_, err := dto.ToTransferFundsCommand()

		assert.ErrorIs(t, err, pkg.ErrNegativeOrZeroAmount)
	})
}

func TestMockTransferDTO_ToReceiveFundsTransferCommand(t *testing.T) {
	t.Run("It should return a valid command when all fields are valid", func(t *testing.T) {
		transferID := valueobject.GenerateID().ToString()
		fromWalletID := valueobject.GenerateID().ToString()
		dto := MockTransferDTO{
			AmountInCents: 300,
			WalletID:      valueobject.GenerateID().ToString(),
			TransferID:    transferID,
			FromWalletID:  fromWalletID,
		}

		command, err := dto.ToReceiveFundsTransferCommand()

		assert.NoError(t, err)
		assert.Equal(t, 300, command.Amount.GetAmount())
		assert.Equal(t, transferID, command.TransferID.ToString())
		assert.Equal(t, fromWalletID, command.FromWalletID.ToString())
		assert.False(t, command.Timestamp.IsZero())
	})

	t.Run("It should return an error when transfer_id is not a valid UUID", func(t *testing.T) {
		dto := MockTransferDTO{
			AmountInCents: 100,
			WalletID:      valueobject.GenerateID().ToString(),
			TransferID:    "not-a-uuid",
			FromWalletID:  valueobject.GenerateID().ToString(),
		}

		_, err := dto.ToReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when from_wallet_id is not a valid UUID", func(t *testing.T) {
		dto := MockTransferDTO{
			AmountInCents: 100,
			WalletID:      valueobject.GenerateID().ToString(),
			TransferID:    valueobject.GenerateID().ToString(),
			FromWalletID:  "not-a-uuid",
		}

		_, err := dto.ToReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrInvalidUUID)
	})

	t.Run("It should return an error when amount_in_cents is zero", func(t *testing.T) {
		dto := MockTransferDTO{
			AmountInCents: 0,
			WalletID:      valueobject.GenerateID().ToString(),
			TransferID:    valueobject.GenerateID().ToString(),
			FromWalletID:  valueobject.GenerateID().ToString(),
		}

		_, err := dto.ToReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrNegativeOrZeroAmount)
	})

	t.Run("It should return an error when amount_in_cents is negative", func(t *testing.T) {
		dto := MockTransferDTO{
			AmountInCents: -1,
			WalletID:      valueobject.GenerateID().ToString(),
			TransferID:    valueobject.GenerateID().ToString(),
			FromWalletID:  valueobject.GenerateID().ToString(),
		}

		_, err := dto.ToReceiveFundsTransferCommand()

		assert.ErrorIs(t, err, pkg.ErrNegativeOrZeroAmount)
	})
}
