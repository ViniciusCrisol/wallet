package persistence

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"wallet/wallet-service/internal/domain"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/eventsourcing"
	"wallet/wallet-service/pkg/valueobject"
)

func WalletDomainToIntegrationEvent(event eventsourcing.Event) (eventsourcing.ParsedEvent, error) {
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

func WalletIntegrationToDomainEvent(
	eventBody []byte,
	eventName string,
) (eventsourcing.Event, error) {
	switch eventName {
	case pkg.WalletCreatedEventName:
		var event pkg.WalletCreatedEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			return nil, err
		}
		walletID, err := valueobject.NewID(event.WalletID)
		if err != nil {
			return nil, err
		}
		holderID, err := valueobject.NewID(event.HolderID)
		if err != nil {
			return nil, err
		}
		return domain.WalletCreatedEvent{
			WalletID:  walletID,
			HolderID:  holderID,
			CreatedAt: event.CreatedAt,
			UpdatedAt: event.UpdatedAt,
		}, nil
	case pkg.FundsTransferredEventName:
		var event pkg.FundsTransferredEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			return nil, err
		}
		transferID, err := valueobject.NewID(event.TransferID)
		if err != nil {
			return nil, err
		}
		toWalletID, err := valueobject.NewID(event.ToWalletID)
		if err != nil {
			return nil, err
		}
		fromWalletID, err := valueobject.NewID(event.FromWalletID)
		if err != nil {
			return nil, err
		}
		amount, err := valueobject.NewMoney(event.AmountInCents)
		if err != nil {
			return nil, err
		}
		return domain.FundsTransferredEvent{
			Amount:       amount,
			TransferID:   transferID,
			ToWalletID:   toWalletID,
			FromWalletID: fromWalletID,
			Timestamp:    event.Timestamp,
		}, nil
	case pkg.FundsTransferReceivedEventName:
		var event pkg.FundsTransferReceivedEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			return nil, err
		}
		walletID, err := valueobject.NewID(event.WalletID)
		if err != nil {
			return nil, err
		}
		transferID, err := valueobject.NewID(event.TransferID)
		if err != nil {
			return nil, err
		}
		fromWalletID, err := valueobject.NewID(event.FromWalletID)
		if err != nil {
			return nil, err
		}
		amount, err := valueobject.NewMoney(event.AmountInCents)
		if err != nil {
			return nil, err
		}
		return domain.FundsTransferReceivedEvent{
			Amount:       amount,
			WalletID:     walletID,
			TransferID:   transferID,
			FromWalletID: fromWalletID,
			Timestamp:    event.Timestamp,
		}, nil
	default:
		slog.Error("unknown event type", slog.String("type", eventName), slog.Any("event", string(eventBody)))
		return nil, pkg.ErrUnknownEventType
	}
}
