package uuid

import (
	"log/slog"

	googleuuid "github.com/google/uuid"
)

func NewUUID() string {
	return NewUUIDValue().String()
}

func IsValid(value string) bool {
	_, err := googleuuid.Parse(value)
	return err == nil
}

func NewUUIDValue() googleuuid.UUID {
	uuidV7, err := googleuuid.NewV7()
	if err != nil {
		slog.Error("failed to generate UUID v7, falling back to v4", slog.String("error", err.Error()))
		return googleuuid.New()
	}
	return uuidV7
}
