package consumer

import (
	"context"
	"log/slog"

	"wallet/wallet-service/config"
	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/internal/command/infrastructure/persistence"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/eventsourcing"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type WalletKurrentDBConsumer struct {
	client    *kurrentdb.Client
	esHandler *persistence.WalletKurrentDBESHandler
}

func NewWalletKurrentDBConsumer(
	client *kurrentdb.Client,
	esHandler *persistence.WalletKurrentDBESHandler,
) *WalletKurrentDBConsumer {
	return &WalletKurrentDBConsumer{
		client:    client,
		esHandler: esHandler,
	}
}

func (consumer *WalletKurrentDBConsumer) Start(ctx context.Context) {
	groupName := config.Load().WalletCommandGroupName
	eventsourcing.SubscribeAndConsume(ctx, consumer.client, consumer.handle, groupName)
}

func (consumer *WalletKurrentDBConsumer) handle(
	ctx context.Context,
	eventBody []byte,
	eventName string,
) error {
	switch eventName {
	case pkg.FundsTransferredEventName:
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
		return err
	}
	if !found {
		slog.Error("wallet not found for funds transfer", slog.String("wallet_id", event.ToWalletID.String()))
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
	return consumer.esHandler.Save(ctx, wallet)
}
