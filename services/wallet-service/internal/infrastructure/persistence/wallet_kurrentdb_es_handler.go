package persistence

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"wallet/wallet-service/internal/domain"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/eventsourcing"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type WalletKurrentESHandler struct {
	client *kurrentdb.Client
}

func NewWalletKurrentESHandler(client *kurrentdb.Client) *WalletKurrentESHandler {
	return &WalletKurrentESHandler{
		client: client,
	}
}

func (esHandler *WalletKurrentESHandler) Save(wallet domain.Wallet) error {
	uncommittedEvents := wallet.GetUncommittedEvents()
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

	expectedRevision := wallet.GetVersion() - len(uncommittedEvents)
	var streamState kurrentdb.StreamState
	if expectedRevision < 0 {
		streamState = kurrentdb.NoStream{}
	} else {
		streamState = kurrentdb.StreamRevision{Value: uint64(expectedRevision)}
	}

	if _, err := esHandler.client.AppendToStream(
		context.Background(),
		esHandler.buildStreamName(wallet.GetID()),
		kurrentdb.AppendToStreamOptions{StreamState: streamState}, eventDataList...,
	); err != nil {
		if eventsourcing.IsKurrentDBConcurrencyError(err) {
			slog.Warn(
				"concurrency conflict appending to stream",
				slog.String("wallet_id", wallet.GetID().ToString()),
				slog.String("error", err.Error()),
			)
			return pkg.ErrConflict
		}
		slog.Error(
			"failed to append events to stream",
			slog.String("wallet_id", wallet.GetID().ToString()),
			slog.String("error", err.Error()),
		)
		return err
	}
	wallet.Commit()
	return nil
}

func (esHandler *WalletKurrentESHandler) Find(id valueobject.ID) (domain.Wallet, bool, error) {
	stream, err := esHandler.client.ReadStream(
		context.Background(),
		esHandler.buildStreamName(id),
		kurrentdb.ReadStreamOptions{
			From:      kurrentdb.Start{},
			Direction: kurrentdb.Forwards,
		}, ^uint64(0))
	if err != nil {
		if eventsourcing.IsKurrentDBNotFoundError(err) {
			return domain.Wallet{}, false, nil
		}
		slog.Error(
			"failed to read stream",
			slog.String("error", err.Error()),
			slog.String("stream", esHandler.buildStreamName(id)),
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
				"failed to receive event",
				slog.String("error", err.Error()),
				slog.String("stream", esHandler.buildStreamName(id)),
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

func (esHandler *WalletKurrentESHandler) buildStreamName(id valueobject.ID) string {
	return "wallet-" + id.ToString()
}
