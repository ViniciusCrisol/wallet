package eventsourcing

import (
	"testing"

	"wallet/wallet-service/pkg/domain/valueobject"

	"github.com/stretchr/testify/assert"
)

type stubEvent struct {
	name string
}

func TestAggregateRoot_NewAggregateRoot(t *testing.T) {
	t.Parallel()

	t.Run("It should return an aggregate root with the given ID when a valid ID is provided", func(t *testing.T) {
		t.Parallel()

		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)
		assert.Equal(t, id, ar.ID())
	})

	t.Run("It should initialize version as -1 when a new aggregate root is created", func(t *testing.T) {
		t.Parallel()

		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)
		assert.Equal(t, -1, ar.Version())
	})

	t.Run("It should initialize with no uncommitted events when a new aggregate root is created", func(t *testing.T) {
		t.Parallel()

		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)
		assert.Empty(t, ar.UncommittedEvents())
	})
}

func TestAggregateRoot_Record(t *testing.T) {
	t.Parallel()

	t.Run("It should append the event to uncommitted events when an event is recorded", func(t *testing.T) {
		t.Parallel()

		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)
		event := stubEvent{name: "created"}

		ar.Record(event)

		assert.Equal(t, []Event{event}, ar.UncommittedEvents())
	})

	t.Run("It should increment the version by one when an event is recorded", func(t *testing.T) {
		t.Parallel()

		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)

		ar.Record(stubEvent{name: "created"})

		assert.Equal(t, 0, ar.Version())
	})

	t.Run("It should increment version for each event recorded when multiple events are recorded", func(t *testing.T) {
		t.Parallel()

		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)

		ar.Record(stubEvent{name: "created"})
		ar.Record(stubEvent{name: "updated"})
		ar.Record(stubEvent{name: "deleted"})

		assert.Equal(t, 2, ar.Version())
	})

	t.Run("It should accumulate all events in order when multiple events are recorded", func(t *testing.T) {
		t.Parallel()

		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)
		e1 := stubEvent{name: "created"}
		e2 := stubEvent{name: "updated"}

		ar.Record(e1)
		ar.Record(e2)

		assert.Equal(t, []Event{e1, e2}, ar.UncommittedEvents())
	})
}

func TestAggregateRoot_Commit(t *testing.T) {
	t.Parallel()

	t.Run("It should clear all uncommitted events when commit is called", func(t *testing.T) {
		t.Parallel()

		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)
		ar.Record(stubEvent{name: "created"})
		ar.Record(stubEvent{name: "updated"})

		ar.Commit()

		assert.Empty(t, ar.UncommittedEvents())
	})

	t.Run("It should preserve the version after commit when events were previously recorded", func(t *testing.T) {
		t.Parallel()

		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)
		ar.Record(stubEvent{name: "created"})
		ar.Record(stubEvent{name: "updated"})

		ar.Commit()

		assert.Equal(t, 1, ar.Version())
	})

	t.Run("It should have no effect when commit is called on a fresh aggregate root", func(t *testing.T) {
		t.Parallel()

		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)

		ar.Commit()

		assert.Empty(t, ar.UncommittedEvents())
		assert.Equal(t, -1, ar.Version())
	})
}
