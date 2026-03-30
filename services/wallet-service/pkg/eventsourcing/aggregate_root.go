package eventsourcing

import "wallet/wallet-service/pkg/valueobject"

type AggregateRoot struct {
	id                valueobject.ID
	version           int
	uncommittedEvents []Event
}

func NewAggregateRoot(id valueobject.ID) AggregateRoot {
	return AggregateRoot{
		id:                id,
		version:           -1,
		uncommittedEvents: []Event{},
	}
}

func (agg *AggregateRoot) Commit() {
	agg.uncommittedEvents = []Event{}
}

func (agg *AggregateRoot) Record(event Event) {
	agg.version++
	agg.uncommittedEvents = append(agg.uncommittedEvents, event)
}

func (agg *AggregateRoot) IncrementVersion() {
	agg.version++
}

func (agg *AggregateRoot) ID() valueobject.ID {
	return agg.id
}

func (agg *AggregateRoot) Version() int {
	return agg.version
}

func (agg *AggregateRoot) UncommittedEvents() []Event {
	return agg.uncommittedEvents
}
