package domain

import (
	"time"

	valueObject "wallet/wallet-service/pkg/domain/value_object"
)

type CreateWalletCommand struct {
	WalletID  valueObject.ID
	HolderID  valueObject.ID
	Timestamp time.Time
}

type TransferFundsCommand struct {
	Amount     valueObject.Money
	TransferID valueObject.ID
	ToWalletID valueObject.ID
	Timestamp  time.Time
}

type ReceiveFundsTransferCommand struct {
	Amount       valueObject.Money
	TransferID   valueObject.ID
	FromWalletID valueObject.ID
	Timestamp    time.Time
}
