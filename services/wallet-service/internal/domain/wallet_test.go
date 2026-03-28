package domain

import (
	"testing"
	"time"

	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/stretchr/testify/assert"
)

func newMoney(t *testing.T, amountInCents int) valueobject.Money {
	t.Helper()

	money, err := valueobject.NewMoney(amountInCents)
	assert.NoError(t, err)
	return money
}

func newTestWallet(t *testing.T) Wallet {
	t.Helper()

	return NewWallet(CreateWalletCommand{
		WalletID:  valueobject.GenerateID(),
		HolderID:  valueobject.GenerateID(),
		Timestamp: time.Now(),
	})
}

func newTestWalletWithBalance(t *testing.T, amountInCents int) Wallet {
	t.Helper()

	wallet := newTestWallet(t)
	err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
		Amount:       newMoney(t, amountInCents),
		TransferID:   valueobject.GenerateID(),
		FromWalletID: valueobject.GenerateID(),
		Timestamp:    time.Now(),
	})
	assert.NoError(t, err)
	wallet.Commit()
	return wallet
}

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

		wallet := NewWallet(cmd)

		assert.Equal(t, walletID.ToString(), wallet.GetID().ToString())
		assert.Equal(t, holderID.ToString(), wallet.holderID.ToString())
		assert.Equal(t, now, wallet.GetCreatedAt())
		assert.Equal(t, now, wallet.GetUpdatedAt())
		assert.Equal(t, 0, wallet.balance.GetAmount())
	})

	t.Run("It should set version to zero when wallet is created", func(t *testing.T) {
		wallet := newTestWallet(t)
		assert.Equal(t, 0, wallet.GetVersion())
	})

	t.Run("It should log one uncommitted event when wallet is created", func(t *testing.T) {
		wallet := newTestWallet(t)
		events := wallet.GetUncommittedEvents()

		assert.Len(t, events, 1)
		assert.IsType(t, WalletCreatedEvent{}, events[0])
	})
}

func TestWallet_TransferFunds(t *testing.T) {
	t.Run("It should return an error when balance is insufficient", func(t *testing.T) {
		wallet := newTestWallet(t)

		err := wallet.TransferFunds(TransferFundsCommand{
			Amount:     newMoney(t, 100),
			TransferID: valueobject.GenerateID(),
			ToWalletID: valueobject.GenerateID(),
			Timestamp:  time.Now(),
		})

		assert.ErrorIs(t, err, pkg.ErrInsufficientBalance)
	})

	t.Run("It should deduct balance and log an event when funds are transferred", func(t *testing.T) {
		wallet := newTestWalletWithBalance(t, 500)

		err := wallet.TransferFunds(TransferFundsCommand{
			Amount:     newMoney(t, 300),
			TransferID: valueobject.GenerateID(),
			ToWalletID: valueobject.GenerateID(),
			Timestamp:  time.Now(),
		})

		assert.NoError(t, err)
		events := wallet.GetUncommittedEvents()
		assert.Len(t, events, 1)
		assert.IsType(t, FundsTransferredEvent{}, events[0])
		assert.Equal(t, 200, wallet.GetBalance().GetAmount())
	})

	t.Run("It should set balance to zero when transferring the entire balance", func(t *testing.T) {
		wallet := newTestWalletWithBalance(t, 500)

		err := wallet.TransferFunds(TransferFundsCommand{
			Amount:     newMoney(t, 500),
			TransferID: valueobject.GenerateID(),
			ToWalletID: valueobject.GenerateID(),
			Timestamp:  time.Now(),
		})

		assert.NoError(t, err)
		assert.Equal(t, 0, wallet.GetBalance().GetAmount())
	})

	t.Run("It should update updatedAt when funds are transferred", func(t *testing.T) {
		wallet := newTestWalletWithBalance(t, 500)
		transferTime := time.Now().Add(time.Hour)

		err := wallet.TransferFunds(TransferFundsCommand{
			Amount:     newMoney(t, 100),
			TransferID: valueobject.GenerateID(),
			ToWalletID: valueobject.GenerateID(),
			Timestamp:  transferTime,
		})

		assert.NoError(t, err)
		assert.Equal(t, transferTime, wallet.GetUpdatedAt())
	})
}

func TestWallet_ReceiveFundsTransfer(t *testing.T) {
	t.Run("It should add funds and log an event when transfer is received", func(t *testing.T) {
		wallet := newTestWallet(t)
		wallet.Commit()

		err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       newMoney(t, 500),
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})

		assert.NoError(t, err)
		events := wallet.GetUncommittedEvents()
		assert.Len(t, events, 1)
		assert.IsType(t, FundsTransferReceivedEvent{}, events[0])
		assert.Equal(t, 500, wallet.GetBalance().GetAmount())
	})

	t.Run("It should return an error when new balance would exceed the limit", func(t *testing.T) {
		wallet := newTestWalletWithBalance(t, 99_999_999)

		err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       newMoney(t, 2),
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})

		assert.ErrorIs(t, err, pkg.ErrBalanceLimitExceeded)
	})

	t.Run("It should accept funds when new balance equals the limit", func(t *testing.T) {
		wallet := newTestWalletWithBalance(t, 99_999_999)

		err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       newMoney(t, 1),
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})

		assert.NoError(t, err)
		assert.Equal(t, MaxBalanceInCents, wallet.GetBalance().GetAmount())
	})

	t.Run("It should update updatedAt when funds are received", func(t *testing.T) {
		wallet := newTestWallet(t)
		wallet.Commit()
		receiveTime := time.Now().Add(time.Hour)

		err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       newMoney(t, 200),
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    receiveTime,
		})

		assert.NoError(t, err)
		assert.Equal(t, receiveTime, wallet.GetUpdatedAt())
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

	t.Run("It should apply FundsTransferredEvent and deduct balance when replayed", func(t *testing.T) {
		wallet := newTestWalletWithBalance(t, 500)
		transferTime := time.Now().Add(time.Hour)

		wallet.Replay(FundsTransferredEvent{
			Amount:       newMoney(t, 200),
			TransferID:   valueobject.GenerateID(),
			ToWalletID:   valueobject.GenerateID(),
			FromWalletID: wallet.GetID(),
			Timestamp:    transferTime,
		})

		assert.Equal(t, 300, wallet.GetBalance().GetAmount())
		assert.Equal(t, transferTime, wallet.GetUpdatedAt())
	})

	t.Run("It should apply FundsTransferReceivedEvent and update balance when replayed", func(t *testing.T) {
		wallet := newTestWallet(t)
		receiveTime := time.Now().Add(time.Hour)

		wallet.Replay(FundsTransferReceivedEvent{
			Amount:       newMoney(t, 750),
			WalletID:     wallet.GetID(),
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    receiveTime,
		})

		assert.Equal(t, 750, wallet.GetBalance().GetAmount())
		assert.Equal(t, receiveTime, wallet.GetUpdatedAt())
	})

	t.Run("It should increment version when event is replayed", func(t *testing.T) {
		var wallet Wallet

		wallet.Replay(WalletCreatedEvent{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})

		assert.Equal(t, 0, wallet.GetVersion())
	})
}

func TestWallet_GetBalance(t *testing.T) {
	t.Run("It should return zero balance when wallet is first created", func(t *testing.T) {
		wallet := newTestWallet(t)
		assert.Equal(t, 0, wallet.GetBalance().GetAmount())
	})

	t.Run("It should return updated balance after funds are received", func(t *testing.T) {
		wallet := newTestWalletWithBalance(t, 250)
		assert.Equal(t, 250, wallet.GetBalance().GetAmount())
	})
}

func TestWallet_GetHolderID(t *testing.T) {
	t.Run("It should return the holder ID provided at creation", func(t *testing.T) {
		holderID := valueobject.GenerateID()
		wallet := NewWallet(CreateWalletCommand{
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
		wallet := NewWallet(CreateWalletCommand{
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
		wallet := NewWallet(CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: now,
		})

		assert.Equal(t, now, wallet.GetUpdatedAt())
	})

	t.Run("It should return the latest timestamp after funds are received", func(t *testing.T) {
		createdAt := time.Now()
		updatedAt := createdAt.Add(time.Hour)

		wallet := NewWallet(CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: createdAt,
		})

		err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       newMoney(t, 100),
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    updatedAt,
		})
		assert.NoError(t, err)

		assert.Equal(t, updatedAt, wallet.GetUpdatedAt())
	})
}
