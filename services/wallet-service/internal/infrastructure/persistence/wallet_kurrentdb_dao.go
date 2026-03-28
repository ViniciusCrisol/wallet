package persistence

import (
	"context"
	"errors"
	"io"

	"wallet/wallet-service/internal/domain"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/eventsourcing"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/google/uuid"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type WalletKurrentdbDAO struct {
	client *kurrentdb.Client
}

func NewWalletKurrentdbDAO(client *kurrentdb.Client) *WalletKurrentdbDAO {
	return &WalletKurrentdbDAO{
		client: client,
	}
}

func (dao *WalletKurrentdbDAO) Save(wallet domain.Wallet) error {
	uncommittedEvents := wallet.GetUncommittedEvents()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	var unprocessedEvents []kurrentdb.EventData
	for _, event := range uncommittedEvents {
		integrationEvent, err := WalletDomainToIntegrationEvent(event)
		if err != nil {
			return err
		}
		data, err := integrationEvent.ToJSON()
		if err != nil {
			return err
		}
		unprocessedEvents = append(unprocessedEvents, kurrentdb.EventData{
			ContentType: kurrentdb.ContentTypeJson,
			EventType:   integrationEvent.Name,
			EventID:     uuid.New(),
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

	if _, err := dao.client.AppendToStream(
		context.Background(),
		dao.buildStreamName(wallet.GetID()),
		kurrentdb.AppendToStreamOptions{StreamState: streamState}, unprocessedEvents...); err != nil {
		if eventsourcing.IsKurrentdbConcurrencyError(err) {
			return pkg.ErrConflict
		}
		return err
	}
	wallet.Commit()
	return nil
}

func (dao *WalletKurrentdbDAO) Find(id valueobject.ID) (domain.Wallet, bool, error) {
	stream, err := dao.client.ReadStream(
		context.Background(),
		dao.buildStreamName(id),
		kurrentdb.ReadStreamOptions{
			From:      kurrentdb.Start{},
			Direction: kurrentdb.Forwards,
		}, ^uint64(0))
	if err != nil {
		if eventsourcing.IsKurrentdbNotFoundError(err) {
			return domain.Wallet{}, false, nil
		}
		return domain.Wallet{}, false, err
	}
	defer stream.Close()

	var wallet domain.Wallet
	for {
		integrationEvent, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			if eventsourcing.IsKurrentdbNotFoundError(err) {
				return domain.Wallet{}, false, nil
			}
			return domain.Wallet{}, false, err
		}

		domainEvent, err := WalletIntegrationToDomainEvent(
			integrationEvent.Event.Data,
			integrationEvent.Event.EventType,
		)
		if err != nil {
			return domain.Wallet{}, false, err
		}
		wallet.Replay(domainEvent)
		wallet.Log(domainEvent)
	}
	wallet.Commit()
	return wallet, true, nil
}

func (dao *WalletKurrentdbDAO) buildStreamName(id valueobject.ID) string {
	return "wallet-" + id.ToString()
}
