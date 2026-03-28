package valueobject

import (
	"cmp"
	"math"

	"wallet/wallet-service/pkg"
)

type Money struct {
	amountInCents int
}

func NewMoney(amountInCents int) (Money, error) {
	if amountInCents <= 0 || amountInCents >= math.MaxInt {
		return Money{}, pkg.ErrInvalidAmount
	}
	return Money{amountInCents: amountInCents}, nil
}

func (money Money) GetAmount() int {
	return money.amountInCents
}

func (money Money) Sum(other Money) (Money, error) {
	return NewMoney(money.amountInCents + other.amountInCents)
}

func (money Money) Sub(other Money) (Money, error) {
	return NewMoney(money.amountInCents - other.amountInCents)
}

func (money Money) Compare(other Money) int {
	return cmp.Compare(money.amountInCents, other.amountInCents)
}
