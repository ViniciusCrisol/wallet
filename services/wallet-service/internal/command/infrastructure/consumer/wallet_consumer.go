package consumer

import (
	"context"
	"fmt"
	"log/slog"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/internal/command/infrastructure/persistence"
	"wallet/wallet-service/pkg/apperr"
	"wallet/wallet-service/pkg/eventsourcing"
	"wallet/wallet-service/pkg/integrationevent"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type WalletKurrentDBConsumer struct {
	client    *kurrentdb.Client
	esHandler *persistence.WalletKurrentDBESHandler
	groupName string
}

func NewWalletKurrentDBConsumer(
	client *kurrentdb.Client,
	esHandler *persistence.WalletKurrentDBESHandler,
	groupName string,
) *WalletKurrentDBConsumer {
	return &WalletKurrentDBConsumer{
		client:    client,
		esHandler: esHandler,
		groupName: groupName,
	}
}

func (consumer *WalletKurrentDBConsumer) Start(ctx context.Context) {
	eventsourcing.SubscribeAndConsume(ctx, consumer.client, consumer.handle, consumer.groupName)
}

func (consumer *WalletKurrentDBConsumer) handle(
	ctx context.Context,
	eventBody []byte,
	eventName string,
) error {
	switch eventName {
	case integrationevent.FundsTransferredEventName:
		domainEvent, err := persistence.WalletIntegrationToDomainEvent(eventBody, eventName)
		if err != nil {
			return err
		}
		return consumer.receiveFundsTransfer(ctx, domainEvent.(domain.FundsTransferredEvent))

	default:
		slog.Warn("unhandled event type in wallet consumer", slog.String("type", eventName))
		return nil
	}
}

func (consumer *WalletKurrentDBConsumer) receiveFundsTransfer(ctx context.Context, event domain.FundsTransferredEvent) error {
	wallet, found, err := consumer.esHandler.Find(ctx, event.ToWalletID)
	if err != nil {
		return fmt.Errorf("receiving transfer %s for wallet %s: %w", event.TransferID, event.ToWalletID, err)
	}
	if !found {
		return fmt.Errorf("receiving transfer %s: %w", event.TransferID, apperr.ErrWalletNotFound)
	}
	command := domain.ReceiveFundsTransferCommand{
		Amount:       event.Amount,
		TransferID:   event.TransferID,
		FromWalletID: event.FromWalletID,
		Timestamp:    event.Timestamp,
	}
	if err := wallet.ReceiveFundsTransfer(command); err != nil {
		return fmt.Errorf("receiving transfer %s for wallet %s: %w", event.TransferID, event.ToWalletID, err)
	}
	if err := consumer.esHandler.Save(ctx, wallet); err != nil {
		return fmt.Errorf("receiving transfer %s for wallet %s: %w", event.TransferID, event.ToWalletID, err)
	}
	return nil
}
