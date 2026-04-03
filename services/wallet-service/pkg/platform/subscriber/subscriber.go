package subscriber

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"wallet/wallet-service/pkg"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

const (
	maxDelay  = time.Minute
	baseDelay = time.Second
)

type EventHandler func(ctx context.Context, eventBody []byte, eventName string) error

func SubscribeAndConsume(
	ctx context.Context,
	group string,
	client *kurrentdb.Client,
	eventHandler EventHandler,
) {
	delay := baseDelay
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if ctx.Err() != nil {
				return
			}
			err := consumeSubscription(ctx, group, client, eventHandler)
			if !errors.Is(err, pkg.ErrSubscriptionFailed) {
				delay = baseDelay
			}
			slog.Warn("subscription ended, reconnecting", slog.String("group", group), slog.String("error", err.Error()))
			delay = min(delay*2, maxDelay)
			time.Sleep(delay)
		}
	}
}

func consumeSubscription(
	ctx context.Context,
	group string,
	client *kurrentdb.Client,
	eventHandler EventHandler,
) error {
	options := kurrentdb.SubscribeToPersistentSubscriptionOptions{}
	subscription, err := client.SubscribeToPersistentSubscriptionToAll(ctx, group, options)
	if err != nil {
		slog.Error("failed to subscribe",
			slog.String("group", group),
			slog.String("error", err.Error()))
		return pkg.ErrSubscriptionFailed
	}
	defer subscription.Close()

	for {
		msg := subscription.Recv()
		if msg.SubscriptionDropped != nil {
			slog.Warn("subscription dropped",
				slog.String("group", group),
				slog.String("error", msg.SubscriptionDropped.Error.Error()))
			return msg.SubscriptionDropped.Error
		}
		if msg.EventAppeared == nil || msg.EventAppeared.Event == nil || msg.EventAppeared.Event.Event == nil {
			continue
		}

		event := msg.EventAppeared.Event
		if err := eventHandler(ctx, event.Event.Data, event.Event.EventType); err != nil {
			nack := kurrentdb.NackActionRetry
			if pkg.IsPermanentError(err) {
				nack = kurrentdb.NackActionPark
			}
			if err := subscription.Nack(err.Error(), nack, event); err != nil {
				slog.Error("failed to nack event", slog.String("error", err.Error()))
			}
			continue
		}
		if err := subscription.Ack(event); err != nil {
			slog.Error("failed to ack event", slog.String("error", err.Error()))
			return err
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

func IsKurrentDBAlreadyExistsError(err error) bool {
	kurrentErr, ok := kurrentdb.FromError(err)
	return !ok && kurrentErr.Code() == kurrentdb.ErrorCodeResourceAlreadyExists
}
