package persistence

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/pkg/apperr"
	"wallet/wallet-service/pkg/eventsourcing"
	"wallet/wallet-service/pkg/integrationevent"
	"wallet/wallet-service/pkg/valueobject"
)

func walletDomainToIntegrationEvent(event eventsourcing.Event) (eventsourcing.ParsedEvent, error) {
	switch e := event.(type) {
	case domain.WalletCreatedEvent:
		return eventsourcing.ParsedEvent{
			Body: integrationevent.WalletCreatedEvent{
				WalletID:  e.WalletID.String(),
				HolderID:  e.HolderID.String(),
				CreatedAt: e.CreatedAt,
				UpdatedAt: e.UpdatedAt,
			},
			Name: integrationevent.WalletCreatedEventName,
		}, nil
	case domain.FundsTransferredEvent:
		return eventsourcing.ParsedEvent{
			Body: integrationevent.FundsTransferredEvent{
				TransferID:    e.TransferID.String(),
				ToWalletID:    e.ToWalletID.String(),
				FromWalletID:  e.FromWalletID.String(),
				AmountInCents: e.Amount.Amount(),
				Timestamp:     e.Timestamp,
			},
			Name: integrationevent.FundsTransferredEventName,
		}, nil
	case domain.FundsTransferReceivedEvent:
		return eventsourcing.ParsedEvent{
			Body: integrationevent.FundsTransferReceivedEvent{
				WalletID:      e.WalletID.String(),
				TransferID:    e.TransferID.String(),
				FromWalletID:  e.FromWalletID.String(),
				AmountInCents: e.Amount.Amount(),
				Timestamp:     e.Timestamp,
			},
			Name: integrationevent.FundsTransferReceivedEventName,
		}, nil
	default:
		slog.Error("unknown domain event type", slog.String("type", fmt.Sprintf("%T", event)), slog.Any("event", event))
		return eventsourcing.ParsedEvent{}, fmt.Errorf("%w: %T", apperr.ErrUnknownEventType, event)
	}
}

func WalletIntegrationToDomainEvent(
	eventBody []byte,
	eventName string,
) (eventsourcing.Event, error) {
	switch eventName {
	case integrationevent.WalletCreatedEventName:
		return walletCreatedToDomain(eventBody)
	case integrationevent.FundsTransferredEventName:
		return fundsTransferredToDomain(eventBody)
	case integrationevent.FundsTransferReceivedEventName:
		return fundsTransferReceivedToDomain(eventBody)
	default:
		slog.Error("unknown integration event type", slog.String("type", eventName), slog.Any("event", string(eventBody)))
		return nil, fmt.Errorf("%w: %s", apperr.ErrUnknownEventType, eventName)
	}
}

func walletCreatedToDomain(body []byte) (eventsourcing.Event, error) {
	var event integrationevent.WalletCreatedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("%w: unmarshaling wallet created event: %s", apperr.ErrValidation, err)
	}
	walletID, err := valueobject.NewID(event.WalletID)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet_id %q in wallet created event: %w", event.WalletID, err)
	}
	holderID, err := valueobject.NewID(event.HolderID)
	if err != nil {
		return nil, fmt.Errorf("invalid holder_id %q in wallet created event: %w", event.HolderID, err)
	}
	return domain.WalletCreatedEvent{
		WalletID:  walletID,
		HolderID:  holderID,
		CreatedAt: event.CreatedAt,
		UpdatedAt: event.UpdatedAt,
	}, nil
}

func fundsTransferredToDomain(body []byte) (eventsourcing.Event, error) {
	var event integrationevent.FundsTransferredEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("%w: unmarshaling funds transferred event: %s", apperr.ErrValidation, err)
	}
	transferID, err := valueobject.NewID(event.TransferID)
	if err != nil {
		return nil, fmt.Errorf("invalid transfer_id %q in funds transferred event: %w", event.TransferID, err)
	}
	toWalletID, err := valueobject.NewID(event.ToWalletID)
	if err != nil {
		return nil, fmt.Errorf("invalid to_wallet_id %q in funds transferred event: %w", event.ToWalletID, err)
	}
	fromWalletID, err := valueobject.NewID(event.FromWalletID)
	if err != nil {
		return nil, fmt.Errorf("invalid from_wallet_id %q in funds transferred event: %w", event.FromWalletID, err)
	}
	amount, err := valueobject.NewMoney(event.AmountInCents)
	if err != nil {
		return nil, fmt.Errorf("invalid amount %d in funds transferred event: %w", event.AmountInCents, err)
	}
	return domain.FundsTransferredEvent{
		Amount:       amount,
		TransferID:   transferID,
		ToWalletID:   toWalletID,
		FromWalletID: fromWalletID,
		Timestamp:    event.Timestamp,
	}, nil
}

func fundsTransferReceivedToDomain(body []byte) (eventsourcing.Event, error) {
	var event integrationevent.FundsTransferReceivedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("%w: unmarshaling funds transfer received event: %s", apperr.ErrValidation, err)
	}
	walletID, err := valueobject.NewID(event.WalletID)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet_id %q in funds transfer received event: %w", event.WalletID, err)
	}
	transferID, err := valueobject.NewID(event.TransferID)
	if err != nil {
		return nil, fmt.Errorf("invalid transfer_id %q in funds transfer received event: %w", event.TransferID, err)
	}
	fromWalletID, err := valueobject.NewID(event.FromWalletID)
	if err != nil {
		return nil, fmt.Errorf("invalid from_wallet_id %q in funds transfer received event: %w", event.FromWalletID, err)
	}
	amount, err := valueobject.NewMoney(event.AmountInCents)
	if err != nil {
		return nil, fmt.Errorf("invalid amount %d in funds transfer received event: %w", event.AmountInCents, err)
	}
	return domain.FundsTransferReceivedEvent{
		Amount:       amount,
		WalletID:     walletID,
		TransferID:   transferID,
		FromWalletID: fromWalletID,
		Timestamp:    event.Timestamp,
	}, nil
}
