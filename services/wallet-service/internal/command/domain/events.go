package domain

import (
	"time"

	"wallet/wallet-service/pkg/valueobject"
)

type WalletCreatedEvent struct {
	WalletID  valueobject.ID
	HolderID  valueobject.ID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FundsTransferredEvent struct {
	Amount       valueobject.Money
	TransferID   valueobject.ID
	ToWalletID   valueobject.ID
	FromWalletID valueobject.ID
	Timestamp    time.Time
}

type FundsTransferReceivedEvent struct {
	Amount       valueobject.Money
	WalletID     valueobject.ID
	TransferID   valueobject.ID
	FromWalletID valueobject.ID
	Timestamp    time.Time
}
