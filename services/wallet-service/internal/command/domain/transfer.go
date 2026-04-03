package domain

import (
	"time"

	"wallet/wallet-service/pkg/domain/valueobject"
)

const (
	TransferKindOutgoing = "outgoing"
	TransferKindIncoming = "incoming"
)

type Transfer struct {
	id        valueobject.ID
	kind      string
	amount    valueobject.Money
	timestamp time.Time
}

func NewOutgoingTransfer(id valueobject.ID, amount valueobject.Money, timestamp time.Time) Transfer {
	return NewTransfer(id, TransferKindOutgoing, amount, timestamp)
}

func NewIncomingTransfer(id valueobject.ID, amount valueobject.Money, timestamp time.Time) Transfer {
	return NewTransfer(id, TransferKindIncoming, amount, timestamp)
}

func NewTransfer(
	id valueobject.ID,
	kind string,
	amount valueobject.Money,
	timestamp time.Time,
) Transfer {
	return Transfer{
		id:        id,
		kind:      kind,
		amount:    amount,
		timestamp: timestamp,
	}
}

func (transfer *Transfer) ID() valueobject.ID {
	return transfer.id
}

func (transfer *Transfer) Kind() string {
	return transfer.kind
}

func (transfer *Transfer) Amount() valueobject.Money {
	return transfer.amount
}

func (transfer *Transfer) Timestamp() time.Time {
	return transfer.timestamp
}
