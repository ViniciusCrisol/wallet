package domain

import (
	"fmt"
	"log/slog"
	"time"

	appErr "wallet/wallet-service/pkg/app_err"
	eventSourcing "wallet/wallet-service/pkg/domain/event_sourcing"
	valueObject "wallet/wallet-service/pkg/domain/value_object"
)

const MaxBalanceInCents = 100_000_000

type Wallet struct {
	eventSourcing.AggregateRoot
	balance   valueObject.Money
	holderID  valueObject.ID
	transfers []Transfer
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
		slog.Warn("insufficient balance",
			slog.String("wallet_id", wallet.ID().String()),
			slog.Int("balance_in_cents", wallet.balance.Amount()),
			slog.Int("amount_in_cents", command.Amount.Amount()))
		return appErr.ErrInsufficientBalance
	}

	event := FundsTransferredEvent{
		Amount:       command.Amount,
		TransferID:   command.TransferID,
		ToWalletID:   command.ToWalletID,
		FromWalletID: wallet.ID(),
		Category:     command.Category,
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
	maxBalance, err := valueObject.NewMoney(MaxBalanceInCents)
	if err != nil {
		return err
	}
	if newBalance.Compare(maxBalance) > 0 {
		slog.Warn("balance limit would be exceeded",
			slog.String("wallet_id", wallet.ID().String()),
			slog.Int("amount_in_cents", command.Amount.Amount()),
			slog.Int("current_balance_in_cents", wallet.balance.Amount()))
		return appErr.ErrBalanceLimitExceeded
	}

	event := FundsTransferReceivedEvent{
		Amount:       command.Amount,
		WalletID:     wallet.ID(),
		TransferID:   command.TransferID,
		FromWalletID: command.FromWalletID,
		Category:     command.Category,
		Timestamp:    command.Timestamp,
	}
	if err := wallet.applyFundsTransferReceived(event); err != nil {
		return err
	}
	wallet.Record(event)
	return nil
}

func (wallet *Wallet) Replay(event eventSourcing.Event) error {
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
		slog.Error("unknown event type", slog.String("event_type", fmt.Sprintf("%T", event)), slog.Any("event", event))
		return appErr.ErrUnknownEventType
	}
	wallet.IncrementVersion()
	return nil
}

func (wallet *Wallet) applyWalletCreated(event WalletCreatedEvent) {
	wallet.AggregateRoot = eventSourcing.NewAggregateRoot(event.WalletID)
	wallet.holderID = event.HolderID
	wallet.createdAt = event.CreatedAt
	wallet.updatedAt = event.UpdatedAt
}

func (wallet *Wallet) applyFundsTransferred(event FundsTransferredEvent) error {
	if wallet.hasTransfer(event.TransferID) {
		return appErr.ErrDuplicateTransfer
	}
	balance, err := wallet.balance.Sub(event.Amount)
	if err != nil {
		return err
	}
	wallet.balance = balance
	wallet.updatedAt = event.Timestamp
	category := event.Category
	if category == "" {
		category = CategoryUnclassified
	}
	wallet.transfers = append(wallet.transfers, NewOutgoingTransfer(event.TransferID, event.Amount, category, event.Timestamp))
	return nil
}

func (wallet *Wallet) applyFundsTransferReceived(event FundsTransferReceivedEvent) error {
	if wallet.hasTransfer(event.TransferID) {
		return appErr.ErrDuplicateTransfer
	}
	balance, err := wallet.balance.Sum(event.Amount)
	if err != nil {
		return err
	}
	wallet.balance = balance
	wallet.updatedAt = event.Timestamp
	category := event.Category
	if category == "" {
		category = CategoryUnclassified
	}
	wallet.transfers = append(wallet.transfers, NewIncomingTransfer(event.TransferID, event.Amount, category, event.Timestamp))
	return nil
}

func (wallet *Wallet) hasTransfer(transferID valueObject.ID) bool {
	for _, transfer := range wallet.transfers {
		if transfer.ID().Equals(transferID) {
			return true
		}
	}
	return false
}

func (wallet *Wallet) Balance() valueObject.Money {
	return wallet.balance
}

func (wallet *Wallet) HolderID() valueObject.ID {
	return wallet.holderID
}

func (wallet *Wallet) CreatedAt() time.Time {
	return wallet.createdAt
}

func (wallet *Wallet) UpdatedAt() time.Time {
	return wallet.updatedAt
}
