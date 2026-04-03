package persistence

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"math"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/domain/valueobject"
	"wallet/wallet-service/pkg/platform/subscriber"
	"wallet/wallet-service/pkg/platform/uuid"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type WalletKurrentDBESHandler struct {
	client *kurrentdb.Client
}

func NewWalletKurrentDBESHandler(client *kurrentdb.Client) *WalletKurrentDBESHandler {
	return &WalletKurrentDBESHandler{
		client: client,
	}
}

func (handler *WalletKurrentDBESHandler) Save(ctx context.Context, wallet domain.Wallet) error {
	uncommittedEvents := wallet.UncommittedEvents()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	var events []kurrentdb.EventData
	for _, event := range uncommittedEvents {
		integrationEvent, err := walletDomainToIntegrationEvent(event)
		if err != nil {
			return err
		}
		data, err := integrationEvent.ToJSON()
		if err != nil {
			return err
		}
		events = append(events, kurrentdb.EventData{
			ContentType: kurrentdb.ContentTypeJson,
			EventType:   integrationEvent.Name,
			EventID:     uuid.NewUUIDValue(),
			Data:        data,
		})
	}

	expectedRevision := wallet.Version() - len(uncommittedEvents)
	var streamState kurrentdb.StreamState
	if expectedRevision < 0 {
		streamState = kurrentdb.NoStream{}
	} else {
		streamState = kurrentdb.StreamRevision{Value: uint64(expectedRevision)}
	}

	if _, err := handler.client.AppendToStream(
		ctx,
		handler.buildStreamName(wallet.ID()),
		kurrentdb.AppendToStreamOptions{StreamState: streamState}, events...,
	); err != nil {
		if subscriber.IsKurrentDBConcurrencyError(err) {
			slog.Warn("optimistic concurrency conflict on wallet stream",
				slog.String("wallet_id", wallet.ID().String()),
				slog.String("error", err.Error()))
			return pkg.ErrConflict
		}
		slog.Error("failed to append events to wallet stream",
			slog.String("wallet_id", wallet.ID().String()),
			slog.String("error", err.Error()))
		return err
	}
	wallet.Commit()
	return nil
}

func (handler *WalletKurrentDBESHandler) Find(ctx context.Context, id valueobject.ID) (domain.Wallet, bool, error) {
	streamName := handler.buildStreamName(id)

	stream, err := handler.client.ReadStream(
		ctx,
		streamName,
		kurrentdb.ReadStreamOptions{
			From:      kurrentdb.Start{},
			Direction: kurrentdb.Forwards,
		}, math.MaxUint64)
	if err != nil {
		if subscriber.IsKurrentDBNotFoundError(err) {
			slog.Info("wallet stream not found", slog.String("wallet_id", id.String()))
			return domain.Wallet{}, false, nil
		}
		slog.Error("failed to read wallet stream",
			slog.String("wallet_id", id.String()),
			slog.String("error", err.Error()))
		return domain.Wallet{}, false, err
	}
	defer stream.Close()

	var wallet domain.Wallet
	for {
		resolvedEvent, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			if subscriber.IsKurrentDBNotFoundError(err) {
				slog.Info("wallet stream not found", slog.String("wallet_id", id.String()))
				return domain.Wallet{}, false, nil
			}
			slog.Error("failed to read event from wallet stream",
				slog.String("wallet_id", id.String()),
				slog.String("error", err.Error()))
			return domain.Wallet{}, false, err
		}
		domainEvent, err := WalletIntegrationToDomainEvent(
			resolvedEvent.Event.Data,
			resolvedEvent.Event.EventType,
		)
		if err != nil {
			slog.Error("failed to map integration event to domain event",
				slog.String("event_type", resolvedEvent.Event.EventType),
				slog.String("wallet_id", id.String()),
				slog.String("error", err.Error()))
			return domain.Wallet{}, false, err
		}
		if err := wallet.Replay(domainEvent); err != nil {
			slog.Error("failed to replay domain event on wallet",
				slog.String("event_type", resolvedEvent.Event.EventType),
				slog.String("wallet_id", id.String()),
				slog.String("error", err.Error()))
			return domain.Wallet{}, false, err
		}
	}
	wallet.Commit()
	return wallet, true, nil
}

func (handler *WalletKurrentDBESHandler) buildStreamName(id valueobject.ID) string {
	return "wallet-" + id.String()
}
