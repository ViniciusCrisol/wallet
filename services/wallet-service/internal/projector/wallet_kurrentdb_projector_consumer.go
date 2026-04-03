package projector

import (
	"context"
	"encoding/json"
	"log/slog"

	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/platform/integrationevent"
	"wallet/wallet-service/pkg/platform/subscriber"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type WalletKurrentDBProjectorConsumer struct {
	group         string
	client        *kurrentdb.Client
	projectionDAO *WalletMySQLProjectionDAO
}

func NewWalletKurrentDBProjectorConsumer(
	group string,
	client *kurrentdb.Client,
	projectionDAO *WalletMySQLProjectionDAO,
) *WalletKurrentDBProjectorConsumer {
	return &WalletKurrentDBProjectorConsumer{
		group:         group,
		client:        client,
		projectionDAO: projectionDAO,
	}
}

func (consumer *WalletKurrentDBProjectorConsumer) Start(ctx context.Context) {
	subscriber.SubscribeAndConsume(ctx, consumer.group, consumer.client, consumer.handle)
}

func (consumer *WalletKurrentDBProjectorConsumer) handle(ctx context.Context, eventBody []byte, eventName string) error {
	switch eventName {
	case integrationevent.WalletCreatedEventName:
		var event integrationevent.WalletCreatedEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			slog.Error("failed to unmarshal wallet created event",
				slog.String("event_name", eventName),
				slog.String("error", err.Error()))
			return pkg.ErrUnprocessableEntity
		}
		return consumer.projectionDAO.CreateWallet(ctx, event)

	case integrationevent.FundsTransferredEventName:
		var event integrationevent.FundsTransferredEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			slog.Error("failed to unmarshal funds transferred event",
				slog.String("event_name", eventName),
				slog.String("error", err.Error()))
			return pkg.ErrUnprocessableEntity
		}
		return consumer.projectionDAO.ApplyFundsTransferred(ctx, event)

	case integrationevent.FundsTransferReceivedEventName:
		var event integrationevent.FundsTransferReceivedEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			slog.Error("failed to unmarshal funds transfer received event",
				slog.String("event_name", eventName),
				slog.String("error", err.Error()))
			return pkg.ErrUnprocessableEntity
		}
		return consumer.projectionDAO.ApplyFundsTransferReceived(ctx, event)

	default:
		slog.Warn("unhandled event type in projection consumer", slog.String("event_type", eventName))
		return nil
	}
}
