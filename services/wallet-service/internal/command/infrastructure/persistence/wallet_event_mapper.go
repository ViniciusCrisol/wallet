package persistence

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"wallet/wallet-service/internal/command/domain"
	appErr "wallet/wallet-service/pkg/app_err"
	eventSourcing "wallet/wallet-service/pkg/domain/event_sourcing"
	valueObject "wallet/wallet-service/pkg/domain/value_object"
	integrationEvent "wallet/wallet-service/pkg/platform/integration_event"
)

func walletDomainToIntegrationEvent(event eventSourcing.Event) (eventSourcing.ParsedEvent, error) {
	switch e := event.(type) {
	case domain.WalletCreatedEvent:
		return eventSourcing.ParsedEvent{
			Body: integrationEvent.WalletCreatedEvent{
				WalletID:  e.WalletID.String(),
				HolderID:  e.HolderID.String(),
				CreatedAt: e.CreatedAt,
				UpdatedAt: e.UpdatedAt,
			},
			Name: integrationEvent.WalletCreatedEventName,
		}, nil

	case domain.FundsTransferredEvent:
		return eventSourcing.ParsedEvent{
			Body: integrationEvent.FundsTransferredEvent{
				TransferID:    e.TransferID.String(),
				ToWalletID:    e.ToWalletID.String(),
				FromWalletID:  e.FromWalletID.String(),
				AmountInCents: e.Amount.Amount(),
				Category:      e.Category,
				Timestamp:     e.Timestamp,
			},
			Name: integrationEvent.FundsTransferredEventName,
		}, nil

	case domain.FundsTransferReceivedEvent:
		return eventSourcing.ParsedEvent{
			Body: integrationEvent.FundsTransferReceivedEvent{
				WalletID:      e.WalletID.String(),
				TransferID:    e.TransferID.String(),
				FromWalletID:  e.FromWalletID.String(),
				AmountInCents: e.Amount.Amount(),
				Category:      e.Category,
				Timestamp:     e.Timestamp,
			},
			Name: integrationEvent.FundsTransferReceivedEventName,
		}, nil

	default:
		slog.Error("unknown domain event type", slog.String("event_type", fmt.Sprintf("%T", event)), slog.Any("event", event))
		return eventSourcing.ParsedEvent{}, appErr.ErrUnknownEventType
	}
}

func WalletIntegrationToDomainEvent(
	eventBody []byte,
	eventName string,
) (eventSourcing.Event, error) {
	switch eventName {
	case integrationEvent.WalletCreatedEventName:
		return walletCreatedToDomain(eventBody)
	case integrationEvent.FundsTransferredEventName:
		return fundsTransferredToDomain(eventBody)
	case integrationEvent.FundsTransferReceivedEventName:
		return fundsTransferReceivedToDomain(eventBody)
	default:
		slog.Error("unknown integration event type", slog.String("event_type", eventName), slog.Any("event", string(eventBody)))
		return nil, appErr.ErrUnknownEventType
	}
}

func walletCreatedToDomain(body []byte) (eventSourcing.Event, error) {
	var event integrationEvent.WalletCreatedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		slog.Error("failed to unmarshal wallet created event", slog.String("error", err.Error()))
		return nil, appErr.ErrUnprocessableEntity
	}
	walletID, err := valueObject.NewID(event.WalletID)
	if err != nil {
		slog.Error("invalid wallet_id in wallet created event",
			slog.String("wallet_id", event.WalletID),
			slog.String("error", err.Error()))
		return nil, err
	}
	holderID, err := valueObject.NewID(event.HolderID)
	if err != nil {
		slog.Error("invalid holder_id in wallet created event",
			slog.String("holder_id", event.HolderID),
			slog.String("error", err.Error()))
		return nil, err
	}
	return domain.WalletCreatedEvent{
		WalletID:  walletID,
		HolderID:  holderID,
		CreatedAt: event.CreatedAt,
		UpdatedAt: event.UpdatedAt,
	}, nil
}

func fundsTransferredToDomain(body []byte) (eventSourcing.Event, error) {
	var event integrationEvent.FundsTransferredEvent
	if err := json.Unmarshal(body, &event); err != nil {
		slog.Error("failed to unmarshal funds transferred event", slog.String("error", err.Error()))
		return nil, appErr.ErrUnprocessableEntity
	}
	transferID, err := valueObject.NewID(event.TransferID)
	if err != nil {
		slog.Error("invalid transfer_id in funds transferred event",
			slog.String("transfer_id", event.TransferID),
			slog.String("error", err.Error()))
		return nil, err
	}
	toWalletID, err := valueObject.NewID(event.ToWalletID)
	if err != nil {
		slog.Error("invalid to_wallet_id in funds transferred event",
			slog.String("to_wallet_id", event.ToWalletID),
			slog.String("error", err.Error()))
		return nil, err
	}
	fromWalletID, err := valueObject.NewID(event.FromWalletID)
	if err != nil {
		slog.Error("invalid from_wallet_id in funds transferred event",
			slog.String("from_wallet_id", event.FromWalletID),
			slog.String("error", err.Error()))
		return nil, err
	}
	amount, err := valueObject.NewMoney(event.AmountInCents)
	if err != nil {
		slog.Error("invalid amount in funds transferred event",
			slog.Int("amount_in_cents", event.AmountInCents),
			slog.String("error", err.Error()))
		return nil, err
	}
	return domain.FundsTransferredEvent{
		Amount:       amount,
		TransferID:   transferID,
		ToWalletID:   toWalletID,
		FromWalletID: fromWalletID,
		Category:     event.Category,
		Timestamp:    event.Timestamp,
	}, nil
}

func fundsTransferReceivedToDomain(body []byte) (eventSourcing.Event, error) {
	var event integrationEvent.FundsTransferReceivedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		slog.Error("failed to unmarshal funds transfer received event", slog.String("error", err.Error()))
		return nil, appErr.ErrUnprocessableEntity
	}
	walletID, err := valueObject.NewID(event.WalletID)
	if err != nil {
		slog.Error("invalid wallet_id in funds transfer received event",
			slog.String("wallet_id", event.WalletID),
			slog.String("error", err.Error()))
		return nil, err
	}
	transferID, err := valueObject.NewID(event.TransferID)
	if err != nil {
		slog.Error("invalid transfer_id in funds transfer received event",
			slog.String("transfer_id", event.TransferID),
			slog.String("error", err.Error()))
		return nil, err
	}
	fromWalletID, err := valueObject.NewID(event.FromWalletID)
	if err != nil {
		slog.Error("invalid from_wallet_id in funds transfer received event",
			slog.String("from_wallet_id", event.FromWalletID),
			slog.String("error", err.Error()))
		return nil, err
	}
	amount, err := valueObject.NewMoney(event.AmountInCents)
	if err != nil {
		slog.Error("invalid amount in funds transfer received event",
			slog.Int("amount_in_cents", event.AmountInCents),
			slog.String("error", err.Error()))
		return nil, err
	}
	return domain.FundsTransferReceivedEvent{
		Amount:       amount,
		WalletID:     walletID,
		TransferID:   transferID,
		FromWalletID: fromWalletID,
		Category:     event.Category,
		Timestamp:    event.Timestamp,
	}, nil
}
