package value_object

import (
	appErr "wallet/wallet-service/pkg/app_err"
	"wallet/wallet-service/pkg/platform/uuid"
)

type ID struct {
	value string
}

func NewID(value string) (ID, error) {
	if !uuid.IsValid(value) {
		return ID{}, appErr.ErrInvalidUUID
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
