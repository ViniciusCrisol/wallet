package consumer

import (
	"context"
	"log/slog"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/internal/command/infrastructure/persistence"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/platform/integrationevent"
	"wallet/wallet-service/pkg/platform/subscriber"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type WalletKurrentDBConsumer struct {
	group     string
	client    *kurrentdb.Client
	esHandler *persistence.WalletKurrentDBESHandler
}

func NewWalletKurrentDBConsumer(
	group string,
	client *kurrentdb.Client,
	esHandler *persistence.WalletKurrentDBESHandler,
) *WalletKurrentDBConsumer {
	return &WalletKurrentDBConsumer{
		group:     group,
		client:    client,
		esHandler: esHandler,
	}
}

func (consumer *WalletKurrentDBConsumer) Start(ctx context.Context) {
	subscriber.SubscribeAndConsume(ctx, consumer.group, consumer.client, consumer.handle)
}

func (consumer *WalletKurrentDBConsumer) handle(ctx context.Context, eventBody []byte, eventName string) error {
	switch eventName {
	case integrationevent.FundsTransferredEventName:
		domainEvent, err := persistence.WalletIntegrationToDomainEvent(eventBody, eventName)
		if err != nil {
			return err
		}
		return consumer.receiveFundsTransfer(ctx, domainEvent.(domain.FundsTransferredEvent))
	default:
		slog.Warn("unhandled event type in wallet consumer", slog.String("event_type", eventName))
		return nil
	}
}

func (consumer *WalletKurrentDBConsumer) receiveFundsTransfer(ctx context.Context, event domain.FundsTransferredEvent) error {
	wallet, found, err := consumer.esHandler.Find(ctx, event.ToWalletID)
	if err != nil {
		return err
	}
	if !found {
		slog.Error("destination wallet not found for funds transfer",
			slog.String("to_wallet_id", event.ToWalletID.String()),
			slog.String("transfer_id", event.TransferID.String()))
		return pkg.ErrWalletNotFound
	}
	command := domain.ReceiveFundsTransferCommand{
		Amount:       event.Amount,
		TransferID:   event.TransferID,
		FromWalletID: event.FromWalletID,
		Timestamp:    event.Timestamp,
	}
	if err := wallet.ReceiveFundsTransfer(command); err != nil {
		return err
	}
	if err := consumer.esHandler.Save(ctx, wallet); err != nil {
		return err
	}
	return nil
}
