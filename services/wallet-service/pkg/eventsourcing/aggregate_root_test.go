package eventsourcing

import (
	"testing"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/stretchr/testify/assert"
)

type stubEvent struct {
	name string
}

func TestAggregateRoot_NewAggregateRoot(t *testing.T) {
	t.Run("It should return an aggregate root with the given ID when a valid ID is provided", func(t *testing.T) {
		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)
		assert.Equal(t, id, ar.GetID())
	})

	t.Run("It should initialize version as -1 when a new aggregate root is created", func(t *testing.T) {
		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)
		assert.Equal(t, -1, ar.GetVersion())
	})

	t.Run("It should initialize with no uncommitted events when a new aggregate root is created", func(t *testing.T) {
		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)
		assert.Empty(t, ar.GetUncommittedEvents())
	})
}

func TestAggregateRoot_Log(t *testing.T) {
	t.Run("It should append the event to uncommitted events when an event is logged", func(t *testing.T) {
		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)
		event := stubEvent{name: "created"}

		ar.Log(event)

		assert.Equal(t, []Event{event}, ar.GetUncommittedEvents())
	})

	t.Run("It should increment the version by one when an event is logged", func(t *testing.T) {
		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)

		ar.Log(stubEvent{name: "created"})

		assert.Equal(t, 0, ar.GetVersion())
	})

	t.Run("It should increment version for each event logged when multiple events are logged", func(t *testing.T) {
		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)

		ar.Log(stubEvent{name: "created"})
		ar.Log(stubEvent{name: "updated"})
		ar.Log(stubEvent{name: "deleted"})

		assert.Equal(t, 2, ar.GetVersion())
	})

	t.Run("It should accumulate all events in order when multiple events are logged", func(t *testing.T) {
		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)
		e1 := stubEvent{name: "created"}
		e2 := stubEvent{name: "updated"}

		ar.Log(e1)
		ar.Log(e2)

		assert.Equal(t, []Event{e1, e2}, ar.GetUncommittedEvents())
	})
}

func TestAggregateRoot_Commit(t *testing.T) {
	t.Run("It should clear all uncommitted events when commit is called", func(t *testing.T) {
		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)
		ar.Log(stubEvent{name: "created"})
		ar.Log(stubEvent{name: "updated"})

		ar.Commit()

		assert.Empty(t, ar.GetUncommittedEvents())
	})

	t.Run("It should preserve the version after commit when events were previously logged", func(t *testing.T) {
		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)
		ar.Log(stubEvent{name: "created"})
		ar.Log(stubEvent{name: "updated"})

		ar.Commit()

		assert.Equal(t, 1, ar.GetVersion())
	})

	t.Run("It should have no effect when commit is called on a fresh aggregate root", func(t *testing.T) {
		id := valueobject.GenerateID()
		ar := NewAggregateRoot(id)

		ar.Commit()

		assert.Empty(t, ar.GetUncommittedEvents())
		assert.Equal(t, -1, ar.GetVersion())
	})
}
