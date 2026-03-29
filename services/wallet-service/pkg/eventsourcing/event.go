package eventsourcing

import (
	"encoding/json"
	"fmt"
)

type Event any

type ParsedEvent struct {
	Body any    `json:"body"`
	Name string `json:"name"`
}

func (event ParsedEvent) ToJSON() ([]byte, error) {
	j, err := json.Marshal(event.Body)
	if err != nil {
		return nil, fmt.Errorf("marshaling event %q to JSON: %w", event.Name, err)
	}
	return j, nil
}
