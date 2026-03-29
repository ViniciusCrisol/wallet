package valueobject

import (
	"cmp"
	"math"

	"wallet/wallet-service/pkg/apperr"
)

type Money struct {
	amountInCents int
}

func NewMoney(amountInCents int) (Money, error) {
	if amountInCents < 0 {
		return Money{}, apperr.ErrNegativeAmount
	}
	if amountInCents >= math.MaxInt {
		return Money{}, apperr.ErrAmountOverflow
	}
	return Money{amountInCents: amountInCents}, nil
}

func (money Money) Amount() int {
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
