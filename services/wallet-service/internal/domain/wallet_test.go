package domain

import (
	"testing"
	"time"

	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/stretchr/testify/assert"
)

func TestNewWallet(t *testing.T) {
	t.Run("It should return a valid wallet when all fields are provided", func(t *testing.T) {
		now := time.Now()
		walletID := valueobject.GenerateID()
		holderID := valueobject.GenerateID()
		cmd := CreateWalletCommand{
			WalletID:  walletID,
			HolderID:  holderID,
			Timestamp: now,
		}

		wallet, err := NewWallet(cmd)

		assert.NoError(t, err)
		assert.Equal(t, walletID.ToString(), wallet.GetID().ToString())
		assert.Equal(t, holderID.ToString(), wallet.holderID.ToString())
		assert.Equal(t, now, wallet.GetCreatedAt())
		assert.Equal(t, now, wallet.GetUpdatedAt())
		assert.Equal(t, 0, wallet.balance.GetAmount())
	})

	t.Run("It should set version to zero when wallet is created", func(t *testing.T) {
		cmd := CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		}

		wallet, _ := NewWallet(cmd)

		assert.Equal(t, 0, wallet.GetVersion())
	})

	t.Run("It should log one uncommitted event when wallet is created", func(t *testing.T) {
		cmd := CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		}

		wallet, _ := NewWallet(cmd)
		events := wallet.GetUncommittedEvents()

		assert.Len(t, events, 1)
		assert.IsType(t, events[0], WalletCreatedEvent{})
	})
}

func TestWallet_TransferFunds(t *testing.T) {
	t.Run("It should return an error when balance is insufficient", func(t *testing.T) {
		wallet, _ := NewWallet(CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})

		amount, _ := valueobject.NewMoney(100)
		err := wallet.TransferFunds(TransferFundsCommand{
			Amount:     amount,
			TransferID: valueobject.GenerateID(),
			ToWalletID: valueobject.GenerateID(),
			Timestamp:  time.Now(),
		})
		assert.ErrorIs(t, err, pkg.ErrInsufficientBalance)
	})

	t.Run("It should deduct balance and log an event when funds are transferred", func(t *testing.T) {
		wallet, _ := NewWallet(CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})
		received, _ := valueobject.NewMoney(500)
		wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       received,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})
		wallet.Commit()

		transferAmount, _ := valueobject.NewMoney(300)
		err := wallet.TransferFunds(TransferFundsCommand{
			Amount:     transferAmount,
			TransferID: valueobject.GenerateID(),
			ToWalletID: valueobject.GenerateID(),
			Timestamp:  time.Now(),
		})

		assert.NoError(t, err)
		events := wallet.GetUncommittedEvents()
		assert.Len(t, events, 1)
		assert.IsType(t, events[0], FundsTransferredEvent{})
	})
}

func TestWallet_ReceiveFundsTransfer(t *testing.T) {
	t.Run("It should add funds and log an event when transfer is received", func(t *testing.T) {
		wallet, _ := NewWallet(CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})
		wallet.Commit()

		amount, _ := valueobject.NewMoney(500)
		err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       amount,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})

		assert.NoError(t, err)
		events := wallet.GetUncommittedEvents()
		assert.Len(t, events, 1)
		assert.IsType(t, events[0], FundsTransferReceivedEvent{})
	})

	t.Run("It should return an error when new balance would exceed the limit", func(t *testing.T) {
		wallet, _ := NewWallet(CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})
		nearMax, _ := valueobject.NewMoney(99_999_999)
		_ = wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       nearMax,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})

		overflow, _ := valueobject.NewMoney(2)
		err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       overflow,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})
		assert.ErrorIs(t, err, pkg.ErrBalanceLimitExceeded)
	})
}

func TestWallet_Replay(t *testing.T) {
	t.Run("It should apply WalletCreatedEvent when replayed", func(t *testing.T) {
		var wallet Wallet
		now := time.Now()
		walletID := valueobject.GenerateID()
		holderID := valueobject.GenerateID()

		wallet.Replay(WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: now,
			UpdatedAt: now,
		})

		assert.Equal(t, walletID.ToString(), wallet.GetID().ToString())
		assert.Equal(t, holderID.ToString(), wallet.holderID.ToString())
		assert.Equal(t, now, wallet.GetCreatedAt())
		assert.Equal(t, now, wallet.GetUpdatedAt())
		assert.Equal(t, 0, wallet.balance.GetAmount())
	})

	t.Run("It should apply FundsTransferReceivedEvent when replayed", func(t *testing.T) {
		wallet, _ := NewWallet(CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})
		nearMax, _ := valueobject.NewMoney(99_999_999)
		wallet.Replay(FundsTransferReceivedEvent{
			Amount:       nearMax,
			WalletID:     wallet.GetID(),
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})

		overflow, _ := valueobject.NewMoney(2)
		err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       overflow,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})
		assert.ErrorIs(t, err, pkg.ErrBalanceLimitExceeded)
	})

	t.Run("It should apply FundsTransferredEvent when replayed", func(t *testing.T) {
		wallet, _ := NewWallet(CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})
		initial, _ := valueobject.NewMoney(500)
		wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       initial,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})
		deducted, _ := valueobject.NewMoney(400)
		wallet.Replay(FundsTransferredEvent{
			Amount:       deducted,
			TransferID:   valueobject.GenerateID(),
			ToWalletID:   valueobject.GenerateID(),
			FromWalletID: wallet.GetID(),
			Timestamp:    time.Now(),
		})

		excess, _ := valueobject.NewMoney(200)
		err := wallet.TransferFunds(TransferFundsCommand{
			Amount:     excess,
			TransferID: valueobject.GenerateID(),
			ToWalletID: valueobject.GenerateID(),
			Timestamp:  time.Now(),
		})
		assert.ErrorIs(t, err, pkg.ErrInsufficientBalance)
	})
}

func TestWallet_GetBalance(t *testing.T) {
	t.Run("It should return zero balance when wallet is first created", func(t *testing.T) {
		wallet, _ := NewWallet(CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})

		assert.Equal(t, 0, wallet.GetBalance().GetAmount())
	})

	t.Run("It should return updated balance after funds are received", func(t *testing.T) {
		wallet, _ := NewWallet(CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})
		amount, _ := valueobject.NewMoney(250)
		_ = wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       amount,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})

		assert.Equal(t, 250, wallet.GetBalance().GetAmount())
	})
}

func TestWallet_GetHolderID(t *testing.T) {
	t.Run("It should return the holder ID provided at creation", func(t *testing.T) {
		holderID := valueobject.GenerateID()
		wallet, _ := NewWallet(CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  holderID,
			Timestamp: time.Now(),
		})

		assert.Equal(t, holderID.ToString(), wallet.GetHolderID().ToString())
	})
}

func TestWallet_GetCreatedAt(t *testing.T) {
	t.Run("It should return the timestamp provided at creation", func(t *testing.T) {
		now := time.Now()
		wallet, _ := NewWallet(CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: now,
		})

		assert.Equal(t, now, wallet.GetCreatedAt())
	})
}

func TestWallet_GetUpdatedAt(t *testing.T) {
	t.Run("It should return the creation timestamp when wallet has not been updated", func(t *testing.T) {
		now := time.Now()
		wallet, _ := NewWallet(CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: now,
		})

		assert.Equal(t, now, wallet.GetUpdatedAt())
	})

	t.Run("It should return the latest timestamp after funds are received", func(t *testing.T) {
		createdAt := time.Now()
		wallet, _ := NewWallet(CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: createdAt,
		})

		updatedAt := createdAt.Add(time.Hour)
		amount, _ := valueobject.NewMoney(100)
		_ = wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       amount,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    updatedAt,
		})

		assert.Equal(t, updatedAt, wallet.GetUpdatedAt())
	})
}
