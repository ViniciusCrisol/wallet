package value_object

import (
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/platform/uuid"
)

type ID struct {
	value string
}

func NewID(value string) (ID, error) {
	if !uuid.IsValid(value) {
		return ID{}, pkg.ErrInvalidUUID
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
