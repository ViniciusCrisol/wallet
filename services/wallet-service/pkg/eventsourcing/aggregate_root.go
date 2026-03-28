package eventsourcing

type Event any

type AggregateRoot struct {
	id                string
	version           int
	uncommittedEvents []Event
}

func NewAggregateRoot(id string) AggregateRoot {
	return AggregateRoot{
		id:                id,
		version:           -1,
		uncommittedEvents: []Event{},
	}
}

func (aggregateRoot *AggregateRoot) Commit() {
	aggregateRoot.uncommittedEvents = []Event{}
}

func (aggregateRoot *AggregateRoot) Log(event Event) {
	aggregateRoot.version++
	aggregateRoot.uncommittedEvents = append(aggregateRoot.uncommittedEvents, event)
}

func (aggregateRoot *AggregateRoot) GetID() string {
	return aggregateRoot.id
}

func (aggregateRoot *AggregateRoot) GetVersion() int {
	return aggregateRoot.version
}

func (aggregateRoot *AggregateRoot) GetUncommittedEvents() []Event {
	return aggregateRoot.uncommittedEvents
}
