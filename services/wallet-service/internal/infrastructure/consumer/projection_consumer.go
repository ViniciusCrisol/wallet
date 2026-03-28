package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"wallet/wallet-service/config"
	"wallet/wallet-service/internal/infrastructure/persistence"
	"wallet/wallet-service/pkg"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type ProjectionConsumer struct {
	client *kurrentdb.Client
	dao    *persistence.WalletMysqlDAO
}

func NewProjectionConsumer(client *kurrentdb.Client, dao *persistence.WalletMysqlDAO) *ProjectionConsumer {
	return &ProjectionConsumer{
		client: client,
		dao:    dao,
	}
}

func (consumer *ProjectionConsumer) Start(ctx context.Context) error {
	subscription, err := consumer.client.SubscribeToPersistentSubscriptionToAll(
		ctx,
		config.Load().WalletProjectionSubscription,
		kurrentdb.SubscribeToPersistentSubscriptionOptions{},
	)
	if err != nil {
		slog.Error("failed to subscribe", slog.String("error", err.Error()))
		return err
	}
	defer subscription.Close()

	for {
		subEvent := subscription.Recv()
		if subEvent.SubscriptionDropped != nil {
			return nil
		}
		if subEvent.EventAppeared == nil ||
			subEvent.EventAppeared.Event == nil ||
			subEvent.EventAppeared.Event.Event == nil {
			continue
		}
		if err := consumer.handle(
			subEvent.EventAppeared.Event.Event.Data,
			subEvent.EventAppeared.Event.Event.EventType,
		); err != nil {
			return err
		}
		if err := subscription.Ack(subEvent.EventAppeared.Event); err != nil {
			return err
		}
	}
}

func (consumer *ProjectionConsumer) handle(
	eventBody []byte,
	eventName string,
) error {
	switch eventName {
	case pkg.WalletCreatedEventName:
		var event pkg.WalletCreatedEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			slog.Error("failed to unmarshal WalletCreatedEvent", slog.String("error", err.Error()))
			return err
		}
		return consumer.dao.CreateWallet(event)

	case pkg.FundsTransferredEventName:
		var event pkg.FundsTransferredEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			slog.Error("failed to unmarshal FundsTransferredEvent", slog.String("error", err.Error()))
			return err
		}
		return consumer.dao.ApplyFundsTransferred(event)

	case pkg.FundsTransferReceivedEventName:
		var event pkg.FundsTransferReceivedEvent
		if err := json.Unmarshal(eventBody, &event); err != nil {
			slog.Error("failed to unmarshal FundsTransferReceivedEvent", slog.String("error", err.Error()))
			return err
		}
		return consumer.dao.ApplyFundsTransferReceived(event)

	default:
		return nil
	}
}
