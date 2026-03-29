package domain

import (
	"fmt"
	"log/slog"
	"time"

	"wallet/wallet-service/pkg/apperr"
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
			slog.String("wallet_id", wallet.ID().String()),
			slog.Int("amount_in_cents", command.Amount.Amount()),
			slog.Int("balance_in_cents", wallet.balance.Amount()),
		)
		return apperr.ErrInsufficientBalance
	}

	event := FundsTransferredEvent{
		Amount:       command.Amount,
		TransferID:   command.TransferID,
		ToWalletID:   command.ToWalletID,
		FromWalletID: wallet.ID(),
		Timestamp:    command.Timestamp,
	}
	if err := wallet.applyFundsTransferred(event); err != nil {
		return err
	}
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
			slog.String("wallet_id", wallet.ID().String()),
			slog.Int("amount_in_cents", command.Amount.Amount()),
			slog.Int("current_balance_in_cents", wallet.balance.Amount()),
		)
		return apperr.ErrBalanceLimitExceeded
	}

	event := FundsTransferReceivedEvent{
		Amount:       command.Amount,
		WalletID:     wallet.ID(),
		TransferID:   command.TransferID,
		FromWalletID: command.FromWalletID,
		Timestamp:    command.Timestamp,
	}
	if err := wallet.applyFundsTransferReceived(event); err != nil {
		return err
	}
	wallet.Record(event)
	return nil
}

func (wallet *Wallet) Replay(event eventsourcing.Event) error {
	switch e := event.(type) {
	case WalletCreatedEvent:
		wallet.applyWalletCreated(e)
	case FundsTransferredEvent:
		if err := wallet.applyFundsTransferred(e); err != nil {
			return err
		}
	case FundsTransferReceivedEvent:
		if err := wallet.applyFundsTransferReceived(e); err != nil {
			return err
		}
	default:
		slog.Error("unknown event type", slog.String("type", fmt.Sprintf("%T", event)), slog.Any("event", event))
		return apperr.ErrUnknownEventType
	}
	wallet.IncrementVersion()
	return nil
}

func (wallet *Wallet) applyWalletCreated(event WalletCreatedEvent) {
	wallet.AggregateRoot = eventsourcing.NewAggregateRoot(event.WalletID)
	wallet.holderID = event.HolderID
	wallet.createdAt = event.CreatedAt
	wallet.updatedAt = event.UpdatedAt
}

func (wallet *Wallet) applyFundsTransferred(event FundsTransferredEvent) error {
	balance, err := wallet.balance.Sub(event.Amount)
	if err != nil {
		return fmt.Errorf("applying FundsTransferredEvent to wallet %s: %w", wallet.ID(), err)
	}
	wallet.balance = balance
	wallet.updatedAt = event.Timestamp
	return nil
}

func (wallet *Wallet) applyFundsTransferReceived(event FundsTransferReceivedEvent) error {
	balance, err := wallet.balance.Sum(event.Amount)
	if err != nil {
		return fmt.Errorf("applying FundsTransferReceivedEvent to wallet %s: %w", wallet.ID(), err)
	}
	wallet.balance = balance
	wallet.updatedAt = event.Timestamp
	return nil
}

func (wallet *Wallet) Balance() valueobject.Money {
	return wallet.balance
}

func (wallet *Wallet) HolderID() valueobject.ID {
	return wallet.holderID
}

func (wallet *Wallet) CreatedAt() time.Time {
	return wallet.createdAt
}

func (wallet *Wallet) UpdatedAt() time.Time {
	return wallet.updatedAt
}
