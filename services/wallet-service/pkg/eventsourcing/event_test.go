package eventsourcing

import (
	"testing"

	"wallet/wallet-service/pkg/integrationevent"

	"github.com/stretchr/testify/assert"
)

func TestParsedEvent_ToJSON(t *testing.T) {
	t.Run("It should return valid JSON when body is a serializable struct", func(t *testing.T) {
		event := ParsedEvent{
			Body: integrationevent.WalletCreatedEvent{
				WalletID: "wallet-123",
				HolderID: "holder-456",
			},
			Name: integrationevent.WalletCreatedEventName,
		}

		result, err := event.ToJSON()

		assert.NoError(t, err)
		assert.JSONEq(t, `{
			"wallet_id": "wallet-123",
			"holder_id": "holder-456",
			"created_at": "0001-01-01T00:00:00Z",
			"updated_at": "0001-01-01T00:00:00Z"
		}`, string(result))
	})

	t.Run("It should return an error when body cannot be marshaled to JSON", func(t *testing.T) {
		event := ParsedEvent{
			Body: make(chan int),
			Name: "some-event",
		}

		result, err := event.ToJSON()

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
