package pkg

import (
	"log/slog"

	"github.com/google/uuid"
)

func NewUUID() string {
	return NewUUIDValue().String()
}

func NewUUIDValue() uuid.UUID {
	uuidV7, err := uuid.NewV7()
	if err != nil {
		slog.Error("failed to generate UUID v7, falling back to v4", slog.String("error", err.Error()))
		return uuid.New()
	}
	return uuidV7
}
