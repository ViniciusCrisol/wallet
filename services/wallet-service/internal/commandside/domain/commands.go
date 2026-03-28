package domain

import (
	"time"

	"wallet/wallet-service/pkg/valueobject"
)

type CreateWalletCommand struct {
	WalletID  valueobject.ID
	HolderID  valueobject.ID
	Timestamp time.Time
}

type TransferFundsCommand struct {
	Amount     valueobject.Money
	TransferID valueobject.ID
	ToWalletID valueobject.ID
	Timestamp  time.Time
}

type ReceiveFundsTransferCommand struct {
	Amount       valueobject.Money
	TransferID   valueobject.ID
	FromWalletID valueobject.ID
	Timestamp    time.Time
}
