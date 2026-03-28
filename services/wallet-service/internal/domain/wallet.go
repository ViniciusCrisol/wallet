package domain

import (
	"fmt"
	"log/slog"
	"time"

	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/eventsourcing"
	"wallet/wallet-service/pkg/valueobject"
)

const MaxBalanceInCents = 100_000_000

type Wallet struct {
	eventsourcing.AggregateRoot

	balance   valueobject.Money
	holderID  valueobject.ID
	createdAt time.Time
	updatedAt time.Time
}

func NewWallet(command CreateWalletCommand) Wallet {
	var wallet Wallet
	event := WalletCreatedEvent{
		WalletID:  command.WalletID,
		HolderID:  command.HolderID,
		CreatedAt: command.Timestamp,
		UpdatedAt: command.Timestamp,
	}
	wallet.applyWalletCreated(event)
	wallet.Record(event)
	return wallet
}

func (wallet *Wallet) TransferFunds(command TransferFundsCommand) error {
	if wallet.balance.Compare(command.Amount) == -1 {
		slog.Warn(
			"insufficient balance",
			slog.String("wallet_id", wallet.GetID().ToString()),
			slog.Int("amount_in_cents", command.Amount.GetAmount()),
			slog.Int("balance_in_cents", wallet.balance.GetAmount()),
		)
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
	wallet.Record(event)
	return nil
}

func (wallet *Wallet) ReceiveFundsTransfer(command ReceiveFundsTransferCommand) error {
	newBalance, err := wallet.balance.Sum(command.Amount)
	if err != nil {
		return err
	}
	maxBalance, err := valueobject.NewMoney(MaxBalanceInCents)
	if err != nil {
		return err
	}
	if newBalance.Compare(maxBalance) > 0 {
		slog.Warn(
			"balance limit would be exceeded",
			slog.String("wallet_id", wallet.GetID().ToString()),
			slog.Int("amount_in_cents", command.Amount.GetAmount()),
			slog.Int("current_balance_in_cents", wallet.balance.GetAmount()),
		)
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
	wallet.Record(event)
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
	default:
		slog.Error("unknown event type", slog.String("type", fmt.Sprintf("%T", event)), slog.Any("event", event))
	}
	wallet.IncrementVersion()
}

func (wallet *Wallet) applyWalletCreated(event WalletCreatedEvent) {
	wallet.AggregateRoot = eventsourcing.NewAggregateRoot(event.WalletID)
	wallet.holderID = event.HolderID
	wallet.createdAt = event.CreatedAt
	wallet.updatedAt = event.UpdatedAt
}

func (wallet *Wallet) applyFundsTransferred(event FundsTransferredEvent) {
	wallet.balance, _ = wallet.balance.Sub(event.Amount)
	wallet.updatedAt = event.Timestamp
}

func (wallet *Wallet) applyFundsTransferReceived(event FundsTransferReceivedEvent) {
	wallet.balance, _ = wallet.balance.Sum(event.Amount)
	wallet.updatedAt = event.Timestamp
}

func (wallet *Wallet) GetBalance() valueobject.Money {
	return wallet.balance
}

func (wallet *Wallet) GetHolderID() valueobject.ID {
	return wallet.holderID
}

func (wallet *Wallet) GetCreatedAt() time.Time {
	return wallet.createdAt
}

func (wallet *Wallet) GetUpdatedAt() time.Time {
	return wallet.updatedAt
}
