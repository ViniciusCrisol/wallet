package domain

import (
	"time"

	valueObject "wallet/wallet-service/pkg/domain/value_object"
)

const (
	TransferKindOutgoing = "outgoing"
	TransferKindIncoming = "incoming"
)

type Transfer struct {
	id        valueObject.ID
	kind      string
	amount    valueObject.Money
	timestamp time.Time
}

func NewOutgoingTransfer(id valueObject.ID, amount valueObject.Money, timestamp time.Time) Transfer {
	return NewTransfer(id, TransferKindOutgoing, amount, timestamp)
}

func NewIncomingTransfer(id valueObject.ID, amount valueObject.Money, timestamp time.Time) Transfer {
	return NewTransfer(id, TransferKindIncoming, amount, timestamp)
}

func NewTransfer(
	id valueObject.ID,
	kind string,
	amount valueObject.Money,
	timestamp time.Time,
) Transfer {
	return Transfer{
		id:        id,
		kind:      kind,
		amount:    amount,
		timestamp: timestamp,
	}
}

func (transfer *Transfer) ID() valueObject.ID {
	return transfer.id
}

func (transfer *Transfer) Kind() string {
	return transfer.kind
}

func (transfer *Transfer) Amount() valueObject.Money {
	return transfer.amount
}

func (transfer *Transfer) Timestamp() time.Time {
	return transfer.timestamp
}
