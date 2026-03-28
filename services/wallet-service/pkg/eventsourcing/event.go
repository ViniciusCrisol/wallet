package eventsourcing

import (
	"encoding/json"
	"log/slog"
)

type Event any

type ParsedEvent struct {
	Body any    `json:"body"`
	Name string `json:"name"`
}

func (event *ParsedEvent) ToJSON() ([]byte, error) {
	j, err := json.Marshal(event.Body)
	if err != nil {
		slog.Error(
			"failed to marshal event to JSON",
			slog.String("error", err.Error()),
			slog.String("name", event.Name),
			slog.Any("event", event),
		)
		return nil, err
	}
	return j, nil
}
