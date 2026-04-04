package domain

import (
	"testing"
	"time"

	"wallet/wallet-service/pkg"
	valueObject "wallet/wallet-service/pkg/domain/value_object"

	"github.com/stretchr/testify/assert"
)

func TestNewWallet(t *testing.T) {
	t.Parallel()

	t.Run("It should return a valid wallet when all fields are provided", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		walletID := valueObject.GenerateID()
		holderID := valueObject.GenerateID()
		cmd := CreateWalletCommand{
			WalletID:  walletID,
			HolderID:  holderID,
			Timestamp: now,
		}

		wallet := NewWallet(cmd)

		assert.Equal(t, walletID.String(), wallet.ID().String())
		assert.Equal(t, holderID.String(), wallet.holderID.String())
		assert.Equal(t, now, wallet.CreatedAt())
		assert.Equal(t, now, wallet.UpdatedAt())
		assert.Equal(t, 0, wallet.balance.Amount())
	})

	t.Run("It should set version to zero when wallet is created", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWallet(t)
		assert.Equal(t, 0, wallet.Version())
	})

	t.Run("It should log one uncommitted event when wallet is created", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWallet(t)
		events := wallet.UncommittedEvents()

		assert.Len(t, events, 1)
		assert.IsType(t, WalletCreatedEvent{}, events[0])
	})
}

func TestWallet_TransferFunds(t *testing.T) {
	t.Parallel()

	t.Run("It should return an error when balance is insufficient", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWallet(t)

		err := wallet.TransferFunds(TransferFundsCommand{
			Amount:     newMoney(t, 100),
			TransferID: valueObject.GenerateID(),
			ToWalletID: valueObject.GenerateID(),
			Category:   CategoryFood,
			Timestamp:  time.Now(),
		})

		assert.ErrorIs(t, err, pkg.ErrInsufficientBalance)
	})

	t.Run("It should deduct balance and log an event when funds are transferred", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWalletWithBalance(t, 500)

		err := wallet.TransferFunds(TransferFundsCommand{
			Amount:     newMoney(t, 300),
			TransferID: valueObject.GenerateID(),
			ToWalletID: valueObject.GenerateID(),
			Category:   CategoryFood,
			Timestamp:  time.Now(),
		})

		assert.NoError(t, err)
		events := wallet.UncommittedEvents()
		assert.Len(t, events, 1)
		assert.IsType(t, FundsTransferredEvent{}, events[0])
		assert.Equal(t, 200, wallet.Balance().Amount())
	})

	t.Run("It should set balance to zero when transferring the entire balance", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWalletWithBalance(t, 500)

		err := wallet.TransferFunds(TransferFundsCommand{
			Amount:     newMoney(t, 500),
			TransferID: valueObject.GenerateID(),
			ToWalletID: valueObject.GenerateID(),
			Category:   CategoryFood,
			Timestamp:  time.Now(),
		})

		assert.NoError(t, err)
		assert.Equal(t, 0, wallet.Balance().Amount())
	})

	t.Run("It should update updatedAt when funds are transferred", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWalletWithBalance(t, 500)
		transferTime := time.Now().Add(time.Hour)

		err := wallet.TransferFunds(TransferFundsCommand{
			Amount:     newMoney(t, 100),
			TransferID: valueObject.GenerateID(),
			ToWalletID: valueObject.GenerateID(),
			Category:   CategoryFood,
			Timestamp:  transferTime,
		})

		assert.NoError(t, err)
		assert.Equal(t, transferTime, wallet.UpdatedAt())
	})

	t.Run("It should return duplicate error when transferring with same TransferID twice", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWalletWithBalance(t, 1000)
		transferID := valueObject.GenerateID()

		err := wallet.TransferFunds(TransferFundsCommand{
			Amount:     newMoney(t, 100),
			TransferID: transferID,
			ToWalletID: valueObject.GenerateID(),
			Category:   CategoryFood,
			Timestamp:  time.Now(),
		})
		assert.NoError(t, err)
		assert.Equal(t, 900, wallet.Balance().Amount())

		err = wallet.TransferFunds(TransferFundsCommand{
			Amount:     newMoney(t, 100),
			TransferID: transferID,
			ToWalletID: valueObject.GenerateID(),
			Category:   CategoryFood,
			Timestamp:  time.Now(),
		})
		assert.ErrorIs(t, err, pkg.ErrDuplicateTransfer)
		assert.Equal(t, 900, wallet.Balance().Amount())
	})
}

func TestWallet_ReceiveFundsTransfer(t *testing.T) {
	t.Parallel()

	t.Run("It should add funds and log an event when transfer is received", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWallet(t)
		wallet.Commit()

		err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       newMoney(t, 500),
			TransferID:   valueObject.GenerateID(),
			FromWalletID: valueObject.GenerateID(),
			Category:     CategoryFood,
			Timestamp:    time.Now(),
		})

		assert.NoError(t, err)
		events := wallet.UncommittedEvents()
		assert.Len(t, events, 1)
		assert.IsType(t, FundsTransferReceivedEvent{}, events[0])
		assert.Equal(t, 500, wallet.Balance().Amount())
	})

	t.Run("It should return an error when new balance would exceed the limit", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWalletWithBalance(t, 99_999_999)

		err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       newMoney(t, 2),
			TransferID:   valueObject.GenerateID(),
			FromWalletID: valueObject.GenerateID(),
			Category:     CategoryFood,
			Timestamp:    time.Now(),
		})

		assert.ErrorIs(t, err, pkg.ErrBalanceLimitExceeded)
	})

	t.Run("It should accept funds when new balance equals the limit", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWalletWithBalance(t, 99_999_999)

		err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       newMoney(t, 1),
			TransferID:   valueObject.GenerateID(),
			FromWalletID: valueObject.GenerateID(),
			Category:     CategoryFood,
			Timestamp:    time.Now(),
		})

		assert.NoError(t, err)
		assert.Equal(t, MaxBalanceInCents, wallet.Balance().Amount())
	})

	t.Run("It should update updatedAt when funds are received", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWallet(t)
		wallet.Commit()
		receiveTime := time.Now().Add(time.Hour)

		err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       newMoney(t, 200),
			TransferID:   valueObject.GenerateID(),
			FromWalletID: valueObject.GenerateID(),
			Category:     CategoryFood,
			Timestamp:    receiveTime,
		})

		assert.NoError(t, err)
		assert.Equal(t, receiveTime, wallet.UpdatedAt())
	})

	t.Run("It should return duplicate error when receiving same TransferID twice", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWallet(t)
		wallet.Commit()
		transferID := valueObject.GenerateID()

		err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       newMoney(t, 500),
			TransferID:   transferID,
			FromWalletID: valueObject.GenerateID(),
			Category:     CategoryFood,
			Timestamp:    time.Now(),
		})
		assert.NoError(t, err)
		assert.Equal(t, 500, wallet.Balance().Amount())

		err = wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       newMoney(t, 500),
			TransferID:   transferID,
			FromWalletID: valueObject.GenerateID(),
			Category:     CategoryFood,
			Timestamp:    time.Now(),
		})
		assert.ErrorIs(t, err, pkg.ErrDuplicateTransfer)
		assert.Equal(t, 500, wallet.Balance().Amount())
	})
}

func TestWallet_Replay(t *testing.T) {
	t.Parallel()

	t.Run("It should apply WalletCreatedEvent when replayed", func(t *testing.T) {
		t.Parallel()

		var wallet Wallet
		now := time.Now()
		walletID := valueObject.GenerateID()
		holderID := valueObject.GenerateID()

		assert.NoError(t, wallet.Replay(WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: now,
			UpdatedAt: now,
		}))

		assert.Equal(t, walletID.String(), wallet.ID().String())
		assert.Equal(t, holderID.String(), wallet.holderID.String())
		assert.Equal(t, now, wallet.CreatedAt())
		assert.Equal(t, now, wallet.UpdatedAt())
		assert.Equal(t, 0, wallet.balance.Amount())
	})

	t.Run("It should apply FundsTransferredEvent and deduct balance when replayed", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWalletWithBalance(t, 500)
		transferTime := time.Now().Add(time.Hour)

		assert.NoError(t, wallet.Replay(FundsTransferredEvent{
			Amount:       newMoney(t, 200),
			TransferID:   valueObject.GenerateID(),
			ToWalletID:   valueObject.GenerateID(),
			FromWalletID: wallet.ID(),
			Category:     CategoryFood,
			Timestamp:    transferTime,
		}))

		assert.Equal(t, 300, wallet.Balance().Amount())
		assert.Equal(t, transferTime, wallet.UpdatedAt())
	})

	t.Run("It should apply FundsTransferReceivedEvent and update balance when replayed", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWallet(t)
		receiveTime := time.Now().Add(time.Hour)

		assert.NoError(t, wallet.Replay(FundsTransferReceivedEvent{
			Amount:       newMoney(t, 750),
			WalletID:     wallet.ID(),
			TransferID:   valueObject.GenerateID(),
			FromWalletID: valueObject.GenerateID(),
			Category:     CategoryFuel,
			Timestamp:    receiveTime,
		}))

		assert.Equal(t, 750, wallet.Balance().Amount())
		assert.Equal(t, receiveTime, wallet.UpdatedAt())
	})

	t.Run("It should increment version when event is replayed", func(t *testing.T) {
		t.Parallel()

		var wallet Wallet

		assert.NoError(t, wallet.Replay(WalletCreatedEvent{
			WalletID:  valueObject.GenerateID(),
			HolderID:  valueObject.GenerateID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}))

		assert.Equal(t, 0, wallet.Version())
	})
}

func TestWallet_Balance(t *testing.T) {
	t.Parallel()

	t.Run("It should return zero balance when wallet is first created", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWallet(t)
		assert.Equal(t, 0, wallet.Balance().Amount())
	})

	t.Run("It should return updated balance after funds are received", func(t *testing.T) {
		t.Parallel()

		wallet := newTestWalletWithBalance(t, 250)
		assert.Equal(t, 250, wallet.Balance().Amount())
	})
}

func TestWallet_HolderID(t *testing.T) {
	t.Parallel()

	t.Run("It should return the holder ID provided at creation", func(t *testing.T) {
		t.Parallel()

		holderID := valueObject.GenerateID()
		wallet := NewWallet(CreateWalletCommand{
			WalletID:  valueObject.GenerateID(),
			HolderID:  holderID,
			Timestamp: time.Now(),
		})

		assert.Equal(t, holderID.String(), wallet.HolderID().String())
	})
}

func TestWallet_CreatedAt(t *testing.T) {
	t.Parallel()

	t.Run("It should return the timestamp provided at creation", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		wallet := NewWallet(CreateWalletCommand{
			WalletID:  valueObject.GenerateID(),
			HolderID:  valueObject.GenerateID(),
			Timestamp: now,
		})

		assert.Equal(t, now, wallet.CreatedAt())
	})
}

func TestWallet_UpdatedAt(t *testing.T) {
	t.Parallel()

	t.Run("It should return the creation timestamp when wallet has not been updated", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		wallet := NewWallet(CreateWalletCommand{
			WalletID:  valueObject.GenerateID(),
			HolderID:  valueObject.GenerateID(),
			Timestamp: now,
		})

		assert.Equal(t, now, wallet.UpdatedAt())
	})

	t.Run("It should return the latest timestamp after funds are received", func(t *testing.T) {
		t.Parallel()

		createdAt := time.Now()
		updatedAt := createdAt.Add(time.Hour)

		wallet := NewWallet(CreateWalletCommand{
			WalletID:  valueObject.GenerateID(),
			HolderID:  valueObject.GenerateID(),
			Timestamp: createdAt,
		})

		err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
			Amount:       newMoney(t, 100),
			TransferID:   valueObject.GenerateID(),
			FromWalletID: valueObject.GenerateID(),
			Timestamp:    updatedAt,
		})
		assert.NoError(t, err)

		assert.Equal(t, updatedAt, wallet.UpdatedAt())
	})
}
