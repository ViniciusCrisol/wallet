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
		return eventsourcing.ParsedEvent{}, apperr.ErrUnknownEventType
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
		return nil, apperr.ErrUnknownEventType
	}
}

func walletCreatedToDomain(body []byte) (eventsourcing.Event, error) {
	var event integrationevent.WalletCreatedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		slog.Error("failed to unmarshal wallet created event", slog.String("error", err.Error()))
		return nil, apperr.ErrUnprocessableEntity
	}
	walletID, err := valueobject.NewID(event.WalletID)
	if err != nil {
		slog.Error("invalid wallet_id in wallet created event", slog.String("wallet_id", event.WalletID), slog.String("error", err.Error()))
		return nil, err
	}
	holderID, err := valueobject.NewID(event.HolderID)
	if err != nil {
		slog.Error("invalid holder_id in wallet created event", slog.String("holder_id", event.HolderID), slog.String("error", err.Error()))
		return nil, err
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
		slog.Error("failed to unmarshal funds transferred event", slog.String("error", err.Error()))
		return nil, apperr.ErrUnprocessableEntity
	}
	transferID, err := valueobject.NewID(event.TransferID)
	if err != nil {
		slog.Error("invalid transfer_id in funds transferred event", slog.String("transfer_id", event.TransferID), slog.String("error", err.Error()))
		return nil, err
	}
	toWalletID, err := valueobject.NewID(event.ToWalletID)
	if err != nil {
		slog.Error("invalid to_wallet_id in funds transferred event", slog.String("to_wallet_id", event.ToWalletID), slog.String("error", err.Error()))
		return nil, err
	}
	fromWalletID, err := valueobject.NewID(event.FromWalletID)
	if err != nil {
		slog.Error("invalid from_wallet_id in funds transferred event", slog.String("from_wallet_id", event.FromWalletID), slog.String("error", err.Error()))
		return nil, err
	}
	amount, err := valueobject.NewMoney(event.AmountInCents)
	if err != nil {
		slog.Error("invalid amount in funds transferred event", slog.Int("amount_in_cents", event.AmountInCents), slog.String("error", err.Error()))
		return nil, err
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
		slog.Error("failed to unmarshal funds transfer received event", slog.String("error", err.Error()))
		return nil, apperr.ErrUnprocessableEntity
	}
	walletID, err := valueobject.NewID(event.WalletID)
	if err != nil {
		slog.Error("invalid wallet_id in funds transfer received event", slog.String("wallet_id", event.WalletID), slog.String("error", err.Error()))
		return nil, err
	}
	transferID, err := valueobject.NewID(event.TransferID)
	if err != nil {
		slog.Error("invalid transfer_id in funds transfer received event", slog.String("transfer_id", event.TransferID), slog.String("error", err.Error()))
		return nil, err
	}
	fromWalletID, err := valueobject.NewID(event.FromWalletID)
	if err != nil {
		slog.Error("invalid from_wallet_id in funds transfer received event", slog.String("from_wallet_id", event.FromWalletID), slog.String("error", err.Error()))
		return nil, err
	}
	amount, err := valueobject.NewMoney(event.AmountInCents)
	if err != nil {
		slog.Error("invalid amount in funds transfer received event", slog.Int("amount_in_cents", event.AmountInCents), slog.String("error", err.Error()))
		return nil, err
	}
	return domain.FundsTransferReceivedEvent{
		Amount:       amount,
		WalletID:     walletID,
		TransferID:   transferID,
		FromWalletID: fromWalletID,
		Timestamp:    event.Timestamp,
	}, nil
}
