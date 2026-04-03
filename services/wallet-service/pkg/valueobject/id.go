package valueobject

import (
	"wallet/wallet-service/pkg/apperr"
	"wallet/wallet-service/pkg/uuid"
)

type ID struct {
	value string
}

func NewID(value string) (ID, error) {
	if !uuid.IsValid(value) {
		return ID{}, apperr.ErrInvalidUUID
	}
	return ID{value: value}, nil
}

func GenerateID() ID {
	return ID{value: uuid.NewUUID()}
}

func (id ID) String() string {
	return id.value
}

func (id ID) Equals(other ID) bool {
	return id.value == other.value
}
