package persistence

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"math"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/pkg/apperr"
	"wallet/wallet-service/pkg/eventsourcing"
	"wallet/wallet-service/pkg/uuid"
	"wallet/wallet-service/pkg/valueobject"

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

func (h *WalletKurrentDBESHandler) Save(ctx context.Context, wallet domain.Wallet) error {
	uncommittedEvents := wallet.UncommittedEvents()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	var eventDataList []kurrentdb.EventData
	for _, event := range uncommittedEvents {
		integrationEvent, err := walletDomainToIntegrationEvent(event)
		if err != nil {
			return err
		}
		data, err := integrationEvent.ToJSON()
		if err != nil {
			return err
		}
		eventDataList = append(eventDataList, kurrentdb.EventData{
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

	if _, err := h.client.AppendToStream(
		ctx,
		h.buildStreamName(wallet.ID()),
		kurrentdb.AppendToStreamOptions{StreamState: streamState}, eventDataList...,
	); err != nil {
		if eventsourcing.IsKurrentDBConcurrencyError(err) {
			slog.Warn(
				"optimistic concurrency conflict on wallet stream",
				slog.String("wallet_id", wallet.ID().String()),
				slog.String("error", err.Error()),
			)
			return apperr.ErrConflict
		}
		slog.Error(
			"failed to append events to wallet stream",
			slog.String("wallet_id", wallet.ID().String()),
			slog.String("error", err.Error()),
		)
		return err
	}
	wallet.Commit()
	return nil
}

func (h *WalletKurrentDBESHandler) Find(ctx context.Context, id valueobject.ID) (domain.Wallet, bool, error) {
	streamName := h.buildStreamName(id)

	stream, err := h.client.ReadStream(
		ctx,
		streamName,
		kurrentdb.ReadStreamOptions{
			From:      kurrentdb.Start{},
			Direction: kurrentdb.Forwards,
		}, math.MaxUint64)
	if err != nil {
		if eventsourcing.IsKurrentDBNotFoundError(err) {
			slog.Error("wallet stream not found", slog.String("wallet_id", id.String()))
			return domain.Wallet{}, false, nil
		}
		slog.Error("failed to read wallet stream", slog.String("wallet_id", id.String()), slog.String("error", err.Error()))
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
			if eventsourcing.IsKurrentDBNotFoundError(err) {
				slog.Error("wallet stream not found", slog.String("wallet_id", id.String()))
				return domain.Wallet{}, false, nil
			}
			slog.Error(
				"failed to read event from wallet stream",
				slog.String("wallet_id", id.String()),
				slog.String("error", err.Error()),
			)
			return domain.Wallet{}, false, err
		}
		domainEvent, err := WalletIntegrationToDomainEvent(
			resolvedEvent.Event.Data,
			resolvedEvent.Event.EventType,
		)
		if err != nil {
			slog.Error(
				"failed to map integration event to domain event",
				slog.String("event_type", resolvedEvent.Event.EventType),
				slog.String("wallet_id", id.String()),
				slog.String("error", err.Error()),
			)
			return domain.Wallet{}, false, err
		}
		if err := wallet.Replay(domainEvent); err != nil {
			slog.Error(
				"failed to replay domain event on wallet",
				slog.String("event_type", resolvedEvent.Event.EventType),
				slog.String("wallet_id", id.String()),
				slog.String("error", err.Error()),
			)
			return domain.Wallet{}, false, err
		}
	}
	wallet.Commit()
	return wallet, true, nil
}

func (h *WalletKurrentDBESHandler) buildStreamName(id valueobject.ID) string {
	return "wallet-" + id.String()
}
