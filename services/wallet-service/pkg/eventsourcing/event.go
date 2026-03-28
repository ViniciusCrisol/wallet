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

func ParsedEventFromJSON(j []byte) (ParsedEvent, error) {
	var event ParsedEvent
	if err := json.Unmarshal(j, &event); err != nil {
		slog.Error(
			"failed to unmarshal event from JSON",
			slog.String("error", err.Error()),
			slog.String("event", string(j)),
		)
		return ParsedEvent{}, err
	}
	return event, nil
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
