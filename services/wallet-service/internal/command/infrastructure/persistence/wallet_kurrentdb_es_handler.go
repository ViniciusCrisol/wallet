package persistence

import (
	"context"
	"errors"
	"fmt"
	"io"
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
			return fmt.Errorf("saving wallet %s: %w", wallet.ID(), err)
		}
		data, err := integrationEvent.ToJSON()
		if err != nil {
			return fmt.Errorf("saving wallet %s: %w", wallet.ID(), err)
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
			return fmt.Errorf("%w: wallet %s was modified concurrently, retry the operation", apperr.ErrConflict, wallet.ID())
		}
		return fmt.Errorf("appending events to wallet %s stream: %w", wallet.ID(), err)
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
			return domain.Wallet{}, false, nil
		}
		return domain.Wallet{}, false, fmt.Errorf("reading wallet %s stream: %w", id, err)
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
			return domain.Wallet{}, false, fmt.Errorf("reading event from wallet %s stream: %w", id, err)
		}
		domainEvent, err := WalletIntegrationToDomainEvent(
			resolvedEvent.Event.Data,
			resolvedEvent.Event.EventType,
		)
		if err != nil {
			return domain.Wallet{}, false, fmt.Errorf("replaying wallet %s: %w", id, err)
		}
		if err := wallet.Replay(domainEvent); err != nil {
			return domain.Wallet{}, false, fmt.Errorf("replaying wallet %s: %w", id, err)
		}
	}
	wallet.Commit()
	return wallet, true, nil
}

func (h *WalletKurrentDBESHandler) buildStreamName(id valueobject.ID) string {
	return "wallet-" + id.String()
}
