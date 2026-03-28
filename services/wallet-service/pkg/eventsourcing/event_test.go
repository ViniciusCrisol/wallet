package eventsourcing

import (
	"testing"

	"wallet/wallet-service/pkg"

	"github.com/stretchr/testify/assert"
)

func TestParsedEventFromJSON(t *testing.T) {
	t.Run("It should return a ParsedEvent when valid JSON is provided", func(t *testing.T) {
		input := []byte(`{
			"body": {
				"wallet_id": "wallet-123",
				"holder_id": "holder-456"
			},
			"name": "wallet:wallet_created_event"
		}`)

		result, err := ParsedEventFromJSON(input)

		assert.NoError(t, err)
		assert.Equal(t, pkg.WalletCreatedEventName, result.Name)
		assert.NotNil(t, result.Body)
	})

	t.Run("It should return an error when invalid JSON is provided", func(t *testing.T) {
		input := []byte(`not valid json`)

		result, err := ParsedEventFromJSON(input)

		assert.Error(t, err)
		assert.Empty(t, result.Name)
		assert.Nil(t, result.Body)
	})
}

func TestParsedEvent_ToJSON(t *testing.T) {
	t.Run("It should return valid JSON when body is a serializable struct", func(t *testing.T) {
		event := ParsedEvent{
			Body: pkg.WalletCreatedEvent{
				WalletID: "wallet-123",
				HolderID: "holder-456",
			},
			Name: pkg.WalletCreatedEventName,
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
