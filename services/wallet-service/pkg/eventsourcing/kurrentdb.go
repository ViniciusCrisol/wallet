package eventsourcing

import (
	"context"
	"log/slog"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type EventHandler func(ctx context.Context, eventBody []byte, eventName string) error

func SubscribeAndConsume(
	ctx context.Context,
	kurrentClient *kurrentdb.Client,
	eventHandler EventHandler,
	groupName string,
) {
	options := kurrentdb.SubscribeToPersistentSubscriptionOptions{}
	subscription, err := kurrentClient.SubscribeToPersistentSubscriptionToAll(ctx, groupName, options)
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
		if err := eventHandler(
			ctx,
			msg.EventAppeared.Event.Event.Data,
			msg.EventAppeared.Event.Event.EventType,
		); err != nil {
			return
		}
		if err := subscription.Ack(msg.EventAppeared.Event); err != nil {
			slog.Error("failed to acknowledge event",
				slog.String("event_type", msg.EventAppeared.Event.Event.EventType),
				slog.String("error", err.Error()),
			)
			return
		}
	}
}

func IsKurrentDBNotFoundError(err error) bool {
	kurrentErr, ok := kurrentdb.FromError(err)
	return !ok && kurrentErr.Code() == kurrentdb.ErrorCodeResourceNotFound
}

func IsKurrentDBConcurrencyError(err error) bool {
	kurrentErr, ok := kurrentdb.FromError(err)
	return !ok && kurrentErr.Code() == kurrentdb.ErrorCodeWrongExpectedVersion
}
