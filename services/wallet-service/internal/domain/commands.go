package domain

import "time"

type CreateWalletCommand struct {
	WalletID  string
	HolderID  string
	Timestamp time.Time
}

type TransferFundsCommand struct {
	Amount     int
	TransferID string
	ToWalletID string
	Timestamp  time.Time
}

type ReceiveFundsTransferCommand struct {
	Amount       int
	TransferID   string
	FromWalletID string
	Timestamp    time.Time
}
