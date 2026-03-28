package valueobject

import (
	"wallet/wallet-service/pkg"

	"github.com/google/uuid"
)

type ID struct {
	value string
}

func NewID(value string) (ID, error) {
	if _, err := uuid.Parse(value); err != nil {
		return ID{}, pkg.ErrInvalidUUID
	}
	return ID{value: value}, nil
}

func GenerateID() ID {
	return ID{value: pkg.NewUUID()}
}

func (id ID) String() string {
	return id.value
}
