package eventsourcing

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"wallet/wallet-service/pkg/apperr"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type EventHandler func(ctx context.Context, eventBody []byte, eventName string) error

func SubscribeAndConsume(
	ctx context.Context,
	kurrentClient *kurrentdb.Client,
	eventHandler EventHandler,
	groupName string,
) {
	const baseDelay = time.Second
	const maxDelay = 30 * time.Second
	delay := baseDelay
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := consumeSubscription(ctx, kurrentClient, eventHandler, groupName)
		if ctx.Err() != nil {
			return
		}
		if !errors.Is(err, apperr.ErrSubscriptionFailed) {
			delay = baseDelay
		}
		slog.Warn(
			"subscription ended, reconnecting",
			slog.Duration("delay", delay),
			slog.String("error", err.Error()),
			slog.String("group_name", groupName),
		)

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}

		delay = min(delay*2, maxDelay)
	}
}

func consumeSubscription(
	ctx context.Context,
	kurrentClient *kurrentdb.Client,
	eventHandler EventHandler,
	groupName string,
) error {
	options := kurrentdb.SubscribeToPersistentSubscriptionOptions{}
	subscription, err := kurrentClient.SubscribeToPersistentSubscriptionToAll(ctx, groupName, options)
	if err != nil {
		slog.Error(
			"failed to subscribe to persistent subscription",
			slog.String("group_name", groupName),
			slog.String("error", err.Error()),
		)
		return apperr.ErrSubscriptionFailed
	}
	defer subscription.Close()

	for {
		msg := subscription.Recv()
		if msg.SubscriptionDropped != nil {
			slog.Warn(
				"subscription dropped",
				slog.String("group_name", groupName),
				slog.String("error", msg.SubscriptionDropped.Error.Error()),
			)
			return msg.SubscriptionDropped.Error
		}
		if msg.EventAppeared == nil ||
			msg.EventAppeared.Event == nil ||
			msg.EventAppeared.Event.Event == nil {
			continue
		}
		event := msg.EventAppeared.Event
		if err := eventHandler(ctx, event.Event.Data, event.Event.EventType); err != nil {
			nackAction := kurrentdb.NackActionRetry
			if apperr.IsPermanentError(err) {
				nackAction = kurrentdb.NackActionPark
			}
			slog.Error(
				"failed to handle event",
				slog.String("error", err.Error()),
				slog.String("group_name", groupName),
				slog.String("event_type", event.Event.EventType),
			)
			if nackErr := subscription.Nack(err.Error(), nackAction, event); nackErr != nil {
				slog.Error(
					"failed to nack event",
					slog.String("error", nackErr.Error()),
					slog.String("event_type", event.Event.EventType),
				)
			}
			continue
		}
		if err := subscription.Ack(event); err != nil {
			slog.Error(
				"failed to acknowledge event",
				slog.String("error", err.Error()),
				slog.String("event_type", event.Event.EventType),
			)
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
