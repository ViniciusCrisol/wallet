package projectionbuilder

import (
	"context"
	"encoding/json"
	"log/slog"

	"wallet/wallet-service/config"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/eventsourcing"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type WalletKurrentDBProjectionConsumer struct {
	client        *kurrentdb.Client
	projectionDAO *WalletMySQLProjectionDAO
}

func NewWalletKurrentDBProjectionConsumer(
	client *kurrentdb.Client,
	projectionDAO *WalletMySQLProjectionDAO,
) *WalletKurrentDBProjectionConsumer {
	return &WalletKurrentDBProjectionConsumer{
		client:        client,
		projectionDAO: projectionDAO,
	}
}

func (consumer *WalletKurrentDBProjectionConsumer) Start(ctx context.Context) {
	groupName := config.Load().WalletProjectionGroupName
	eventsourcing.SubscribeAndConsume(ctx, consumer.client, consumer.handle, groupName)
}

func (consumer *WalletKurrentDBProjectionConsumer) handle(
	ctx context.Context,
	eventBody []byte,
	eventName string,
) error {
	switch eventName {
	case pkg.WalletCreatedEventName:
		var event pkg.WalletCreatedEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			slog.Error("failed to unmarshal wallet created event", slog.String("error", err.Error()))
			return err
		}
		return consumer.projectionDAO.CreateWallet(ctx, event)

	case pkg.FundsTransferredEventName:
		var event pkg.FundsTransferredEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			slog.Error("failed to unmarshal funds transferred event", slog.String("error", err.Error()))
			return err
		}
		return consumer.projectionDAO.ApplyFundsTransferred(ctx, event)

	case pkg.FundsTransferReceivedEventName:
		var event pkg.FundsTransferReceivedEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			slog.Error("failed to unmarshal funds transfer received event", slog.String("error", err.Error()))
			return err
		}
		return consumer.projectionDAO.ApplyFundsTransferReceived(ctx, event)

	default:
		slog.Warn("unhandled event type in projection consumer", slog.String("type", eventName))
		return nil
	}
}
