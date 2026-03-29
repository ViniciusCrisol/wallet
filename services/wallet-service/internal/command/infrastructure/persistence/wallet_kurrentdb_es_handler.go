package persistence

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/eventsourcing"
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

func (esHandler *WalletKurrentDBESHandler) Save(ctx context.Context, wallet domain.Wallet) error {
	uncommittedEvents := wallet.UncommittedEvents()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	var eventDataList []kurrentdb.EventData
	for _, event := range uncommittedEvents {
		integrationEvent, err := WalletDomainToIntegrationEvent(event)
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
			EventID:     pkg.NewUUIDValue(),
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

	if _, err := esHandler.client.AppendToStream(
		ctx,
		esHandler.buildStreamName(wallet.ID()),
		kurrentdb.AppendToStreamOptions{StreamState: streamState}, eventDataList...,
	); err != nil {
		if eventsourcing.IsKurrentDBConcurrencyError(err) {
			slog.Warn(
				"concurrency conflict appending to stream",
				slog.String("wallet_id", wallet.ID().String()),
				slog.String("error", err.Error()),
			)
			return pkg.ErrConflict
		}
		slog.Error(
			"failed to append events to stream",
			slog.String("wallet_id", wallet.ID().String()),
			slog.String("error", err.Error()),
		)
		return err
	}
	wallet.Commit()
	return nil
}

func (esHandler *WalletKurrentDBESHandler) Find(ctx context.Context, id valueobject.ID) (domain.Wallet, bool, error) {
	const readAllEvents = ^uint64(0)
	streamName := esHandler.buildStreamName(id)

	stream, err := esHandler.client.ReadStream(
		ctx,
		streamName,
		kurrentdb.ReadStreamOptions{
			From:      kurrentdb.Start{},
			Direction: kurrentdb.Forwards,
		}, readAllEvents)
	if err != nil {
		if eventsourcing.IsKurrentDBNotFoundError(err) {
			return domain.Wallet{}, false, nil
		}
		slog.Error(
			"failed to read stream",
			slog.String("error", err.Error()),
			slog.String("stream", streamName),
		)
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
				return domain.Wallet{}, false, nil
			}
			slog.Error(
				"failed to read event from stream",
				slog.String("error", err.Error()),
				slog.String("stream", streamName),
			)
			return domain.Wallet{}, false, err
		}
		domainEvent, err := WalletIntegrationToDomainEvent(
			resolvedEvent.Event.Data,
			resolvedEvent.Event.EventType,
		)
		if err != nil {
			return domain.Wallet{}, false, err
		}
		wallet.Replay(domainEvent)
	}
	wallet.Commit()
	return wallet, true, nil
}

func (esHandler *WalletKurrentDBESHandler) buildStreamName(id valueobject.ID) string {
	return "wallet-" + id.String()
}
