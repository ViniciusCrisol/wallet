package integration_event

import "time"

const (
	WalletCreatedEventName         = "wallet:wallet_created_event"
	FundsTransferredEventName      = "wallet:funds_transferred_event"
	FundsTransferReceivedEventName = "wallet:funds_transfer_received_event"
)

type WalletCreatedEvent struct {
	WalletID  string    `json:"wallet_id"`
	HolderID  string    `json:"holder_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FundsTransferredEvent struct {
	TransferID    string    `json:"transfer_id"`
	ToWalletID    string    `json:"to_wallet_id"`
	FromWalletID  string    `json:"from_wallet_id"`
	AmountInCents int       `json:"amount_in_cents"`
	Category      string    `json:"category"`
	Timestamp     time.Time `json:"timestamp"`
}

type FundsTransferReceivedEvent struct {
	WalletID      string    `json:"wallet_id"`
	TransferID    string    `json:"transfer_id"`
	FromWalletID  string    `json:"from_wallet_id"`
	AmountInCents int       `json:"amount_in_cents"`
	Category      string    `json:"category"`
	Timestamp     time.Time `json:"timestamp"`
}
