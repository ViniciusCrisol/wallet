package domain

import (
	"time"

	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/eventsourcing"
)

type Wallet struct {
	balance   int
	holderID  string
	createdAt time.Time
	updatedAt time.Time

	eventsourcing.AggregateRoot
}

func NewWallet(command CreateWalletCommand) (Wallet, error) {
	if command.WalletID == "" {
		return Wallet{}, pkg.ErrInvalidID
	}
	if command.HolderID == "" {
		return Wallet{}, pkg.ErrInvalidID
	}

	var wallet Wallet
	event := WalletCreatedEvent{
		WalletID:  command.WalletID,
		HolderID:  command.HolderID,
		CreatedAt: command.Timestamp,
		UpdatedAt: command.Timestamp,
	}
	wallet.applyWalletCreated(event)
	wallet.Log(event)
	return wallet, nil
}

func (wallet *Wallet) TransferFunds(command TransferFundsCommand) error {
	if command.Amount <= 0 {
		return pkg.ErrInvalidAmount
	}
	if wallet.balance < command.Amount {
		return pkg.ErrInsufficientBalance
	}

	event := FundsTransferredEvent{
		Amount:       command.Amount,
		TransferID:   command.TransferID,
		ToWalletID:   command.ToWalletID,
		FromWalletID: wallet.GetID(),
		Timestamp:    command.Timestamp,
	}
	wallet.applyFundsTransferred(event)
	wallet.Log(event)
	return nil
}

func (wallet *Wallet) ReceiveFundsTransfer(command ReceiveFundsTransferCommand) error {
	if command.Amount <= 0 {
		return pkg.ErrInvalidAmount
	}
	if wallet.balance+command.Amount > 1_000_000 {
		return pkg.ErrBalanceLimitExceeded
	}

	event := FundsTransferReceivedEvent{
		Amount:       command.Amount,
		WalletID:     wallet.GetID(),
		TransferID:   command.TransferID,
		FromWalletID: command.FromWalletID,
		Timestamp:    command.Timestamp,
	}
	wallet.applyFundsTransferReceived(event)
	wallet.Log(event)
	return nil
}

func (wallet *Wallet) Replay(event eventsourcing.Event) {
	switch e := event.(type) {
	case WalletCreatedEvent:
		wallet.applyWalletCreated(e)
	case FundsTransferredEvent:
		wallet.applyFundsTransferred(e)
	case FundsTransferReceivedEvent:
		wallet.applyFundsTransferReceived(e)
	}
}

func (wallet *Wallet) applyWalletCreated(event WalletCreatedEvent) {
	wallet.holderID = event.HolderID
	wallet.createdAt = event.CreatedAt
	wallet.updatedAt = event.UpdatedAt
	wallet.AggregateRoot = eventsourcing.NewAggregateRoot(event.WalletID)
}

func (wallet *Wallet) applyFundsTransferred(event FundsTransferredEvent) {
	wallet.balance -= event.Amount
	wallet.updatedAt = event.Timestamp
}

func (wallet *Wallet) applyFundsTransferReceived(event FundsTransferReceivedEvent) {
	wallet.balance += event.Amount
	wallet.updatedAt = event.Timestamp
}

func (wallet *Wallet) GetBalance() int {
	return wallet.balance
}

func (wallet *Wallet) GetHolderID() string {
	return wallet.holderID
}

func (wallet *Wallet) GetCreatedAt() time.Time {
	return wallet.createdAt
}

func (wallet *Wallet) GetUpdatedAt() time.Time {
	return wallet.updatedAt
}
