package projection

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"wallet/wallet-service/pkg/apperr"
	"wallet/wallet-service/pkg/eventsourcing"
	"wallet/wallet-service/pkg/integrationevent"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type WalletKurrentDBProjectionConsumer struct {
	client        *kurrentdb.Client
	projectionDAO *WalletMySQLProjectionDAO
	groupName     string
}

func NewWalletKurrentDBProjectionConsumer(
	client *kurrentdb.Client,
	projectionDAO *WalletMySQLProjectionDAO,
	groupName string,
) *WalletKurrentDBProjectionConsumer {
	return &WalletKurrentDBProjectionConsumer{
		client:        client,
		projectionDAO: projectionDAO,
		groupName:     groupName,
	}
}

func (consumer *WalletKurrentDBProjectionConsumer) Start(ctx context.Context) {
	eventsourcing.SubscribeAndConsume(ctx, consumer.client, consumer.handle, consumer.groupName)
}

func (consumer *WalletKurrentDBProjectionConsumer) handle(
	ctx context.Context,
	eventBody []byte,
	eventName string,
) error {
	switch eventName {
	case integrationevent.WalletCreatedEventName:
		var event integrationevent.WalletCreatedEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			return fmt.Errorf("%w: unmarshaling %s: %s", apperr.ErrValidation, eventName, err)
		}
		return consumer.projectionDAO.CreateWallet(ctx, event)

	case integrationevent.FundsTransferredEventName:
		var event integrationevent.FundsTransferredEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			return fmt.Errorf("%w: unmarshaling %s: %s", apperr.ErrValidation, eventName, err)
		}
		return consumer.projectionDAO.ApplyFundsTransferred(ctx, event)

	case integrationevent.FundsTransferReceivedEventName:
		var event integrationevent.FundsTransferReceivedEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			return fmt.Errorf("%w: unmarshaling %s: %s", apperr.ErrValidation, eventName, err)
		}
		return consumer.projectionDAO.ApplyFundsTransferReceived(ctx, event)

	default:
		slog.Warn("unhandled event type in projection consumer", slog.String("type", eventName))
		return nil
	}
}
