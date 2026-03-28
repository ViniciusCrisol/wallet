package domain

import "time"

type WalletCreatedEvent struct {
	WalletID  string
	HolderID  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FundsTransferredEvent struct {
	Amount       int
	TransferID   string
	ToWalletID   string
	FromWalletID string
	Timestamp    time.Time
}

type FundsTransferReceivedEvent struct {
	Amount       int
	WalletID     string
	TransferID   string
	FromWalletID string
	Timestamp    time.Time
}
