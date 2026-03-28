package projectionbuilder

import (
	"context"
	"encoding/json"
	"log/slog"

	"wallet/wallet-service/config"
	"wallet/wallet-service/pkg"

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
	subscriptionOptions := kurrentdb.SubscribeToPersistentSubscriptionOptions{}
	subscription, err := consumer.client.SubscribeToPersistentSubscriptionToAll(ctx, groupName, subscriptionOptions)
	if err != nil {
		slog.Error("failed to subscribe to persistent subscription", slog.String("group_name", groupName), slog.String("error", err.Error()))
		return
	}
	defer subscription.Close()

	for {
		msg := subscription.Recv()
		if msg.SubscriptionDropped != nil {
			return
		}
		if msg.EventAppeared == nil ||
			msg.EventAppeared.Event == nil ||
			msg.EventAppeared.Event.Event == nil {
			continue
		}
		if err := consumer.handle(
			msg.EventAppeared.Event.Event.Data,
			msg.EventAppeared.Event.Event.EventType,
		); err != nil {
			return
		}
		if err := subscription.Ack(msg.EventAppeared.Event); err != nil {
			return
		}
	}
}

func (consumer *WalletKurrentDBProjectionConsumer) handle(
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
		return consumer.projectionDAO.CreateWallet(event)

	case pkg.FundsTransferredEventName:
		var event pkg.FundsTransferredEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			slog.Error("failed to unmarshal funds transferred event", slog.String("error", err.Error()))
			return err
		}
		return consumer.projectionDAO.ApplyFundsTransferred(event)

	case pkg.FundsTransferReceivedEventName:
		var event pkg.FundsTransferReceivedEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			slog.Error("failed to unmarshal funds transfer received event", slog.String("error", err.Error()))
			return err
		}
		return consumer.projectionDAO.ApplyFundsTransferReceived(event)

	default:
		slog.Warn("unhandled event type in projection consumer", slog.String("type", eventName))
		return nil
	}
}
