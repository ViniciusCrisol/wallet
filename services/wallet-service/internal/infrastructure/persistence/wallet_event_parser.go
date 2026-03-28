package persistence

import (
	"fmt"
	"log/slog"

	"wallet/wallet-service/internal/domain"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/eventsourcing"
)

func ParseWalletEvent(event eventsourcing.Event) (eventsourcing.ParsedEvent, error) {
	switch e := event.(type) {
	case domain.WalletCreatedEvent:
		return eventsourcing.ParsedEvent{
			Body: pkg.WalletCreatedEvent{
				WalletID:  e.WalletID.ToString(),
				HolderID:  e.HolderID.ToString(),
				CreatedAt: e.CreatedAt,
				UpdatedAt: e.UpdatedAt,
			},
			Name: pkg.WalletCreatedEventName,
		}, nil
	case domain.FundsTransferredEvent:
		return eventsourcing.ParsedEvent{
			Body: pkg.FundsTransferredEvent{
				TransferID:    e.TransferID.ToString(),
				ToWalletID:    e.ToWalletID.ToString(),
				FromWalletID:  e.FromWalletID.ToString(),
				AmountInCents: e.Amount.GetAmount(),
				Timestamp:     e.Timestamp,
			},
			Name: pkg.FundsTransferredEventName,
		}, nil
	case domain.FundsTransferReceivedEvent:
		return eventsourcing.ParsedEvent{
			Body: pkg.FundsTransferReceivedEvent{
				WalletID:      e.WalletID.ToString(),
				TransferID:    e.TransferID.ToString(),
				FromWalletID:  e.FromWalletID.ToString(),
				AmountInCents: e.Amount.GetAmount(),
				Timestamp:     e.Timestamp,
			},
			Name: pkg.FundsTransferReceivedEventName,
		}, nil
	default:
		slog.Error("unknown event type", slog.String("type", fmt.Sprintf("%T", event)), slog.Any("event", event))
		return eventsourcing.ParsedEvent{}, pkg.ErrUnknownEventType
	}
}
