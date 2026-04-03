package eventsourcing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParsedEvent_ToJSON(t *testing.T) {
	t.Parallel()

	t.Run("It should return valid JSON when body is a serializable struct", func(t *testing.T) {
		t.Parallel()

		event := ParsedEvent{
			Body: struct {
				WalletID  string    `json:"wallet_id"`
				HolderID  string    `json:"holder_id"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
			}{
				WalletID: "wallet-123",
				HolderID: "holder-456",
			},
			Name: "wallet:wallet_created_event",
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
		t.Parallel()

		event := ParsedEvent{
			Body: make(chan int),
			Name: "some-event",
		}

		result, err := event.ToJSON()

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
