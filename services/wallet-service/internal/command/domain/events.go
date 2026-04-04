package domain

import (
	"time"

	valueObject "wallet/wallet-service/pkg/domain/value_object"
)

type WalletCreatedEvent struct {
	WalletID  valueObject.ID
	HolderID  valueObject.ID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FundsTransferredEvent struct {
	Amount       valueObject.Money
	TransferID   valueObject.ID
	ToWalletID   valueObject.ID
	FromWalletID valueObject.ID
	Category     string
	Timestamp    time.Time
}

type FundsTransferReceivedEvent struct {
	Amount       valueObject.Money
	WalletID     valueObject.ID
	TransferID   valueObject.ID
	FromWalletID valueObject.ID
	Category     string
	Timestamp    time.Time
}
