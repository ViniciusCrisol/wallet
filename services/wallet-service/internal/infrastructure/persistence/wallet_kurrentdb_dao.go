package persistence

import (
	"context"
	"errors"
	"io"

	"wallet/wallet-service/internal/domain"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/eventsourcing"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

type WalletKurrentDBDAO struct {
	client *kurrentdb.Client
}

func NewWalletKurrentDBDAO(client *kurrentdb.Client) *WalletKurrentDBDAO {
	return &WalletKurrentDBDAO{
		client: client,
	}
}

func (dao *WalletKurrentDBDAO) Save(wallet domain.Wallet) error {
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

	if _, err := dao.client.AppendToStream(
		context.Background(),
		dao.buildStreamName(wallet.GetID()),
		kurrentdb.AppendToStreamOptions{StreamState: streamState}, eventDataList...,
	); err != nil {
		if eventsourcing.IsKurrentDBConcurrencyError(err) {
			return pkg.ErrConflict
		}
		return err
	}
	wallet.Commit()
	return nil
}

func (dao *WalletKurrentDBDAO) Find(id valueobject.ID) (domain.Wallet, bool, error) {
	stream, err := dao.client.ReadStream(
		context.Background(),
		dao.buildStreamName(id),
		kurrentdb.ReadStreamOptions{
			From:      kurrentdb.Start{},
			Direction: kurrentdb.Forwards,
		}, ^uint64(0))
	if err != nil {
		if eventsourcing.IsKurrentDBNotFoundError(err) {
			return domain.Wallet{}, false, nil
		}
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

func (dao *WalletKurrentDBDAO) buildStreamName(id valueobject.ID) string {
	return "wallet-" + id.ToString()
}
