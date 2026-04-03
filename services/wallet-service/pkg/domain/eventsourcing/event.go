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

func (event ParsedEvent) ToJSON() ([]byte, error) {
	j, err := json.Marshal(event.Body)
	if err != nil {
		slog.Error("failed to marshal event to JSON",
			slog.String("event_name", event.Name),
			slog.String("error", err.Error()))
		return nil, err
	}
	return j, nil
}
