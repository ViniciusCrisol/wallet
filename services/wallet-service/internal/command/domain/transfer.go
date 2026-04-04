package domain

import (
	"time"

	valueObject "wallet/wallet-service/pkg/domain/value_object"
)

const (
	TransferKindOutgoing = "outgoing"
	TransferKindIncoming = "incoming"

	CategoryFood          = "food"
	CategoryFuel          = "fuel"
	CategorySports        = "sports"
	CategoryHealth        = "health"
	CategoryTravel        = "travel"
	CategoryEssentials    = "essentials"
	CategoryEntertainment = "entertainment"
	CategoryUnclassified  = "unclassified"
)

func IsValidCategory(category string) bool {
	switch category {
	case CategoryFood,
		CategoryFuel,
		CategorySports,
		CategoryHealth,
		CategoryTravel,
		CategoryEssentials,
		CategoryEntertainment,
		CategoryUnclassified:
		return true
	default:
		return false
	}
}

type Transfer struct {
	id        valueObject.ID
	kind      string
	category  string
	amount    valueObject.Money
	timestamp time.Time
}

func NewOutgoingTransfer(id valueObject.ID, amount valueObject.Money, category string, timestamp time.Time) Transfer {
	return NewTransfer(id, TransferKindOutgoing, amount, category, timestamp)
}

func NewIncomingTransfer(id valueObject.ID, amount valueObject.Money, category string, timestamp time.Time) Transfer {
	return NewTransfer(id, TransferKindIncoming, amount, category, timestamp)
}

func NewTransfer(
	id valueObject.ID,
	kind string,
	amount valueObject.Money,
	category string,
	timestamp time.Time,
) Transfer {
	return Transfer{
		id:        id,
		kind:      kind,
		category:  category,
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

func (transfer *Transfer) Category() string {
	return transfer.category
}

func (transfer *Transfer) Amount() valueObject.Money {
	return transfer.amount
}

func (transfer *Transfer) Timestamp() time.Time {
	return transfer.timestamp
}
