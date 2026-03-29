package valueobject

import (
	"math"
	"testing"

	"wallet/wallet-service/pkg/apperr"

	"github.com/stretchr/testify/assert"
)

func TestNewMoney(t *testing.T) {
	t.Run("It should return a valid Money when amount is positive", func(t *testing.T) {
		money, err := NewMoney(100)
		assert.NoError(t, err)
		assert.Equal(t, 100, money.Amount())
	})

	t.Run("It should return a valid Money when amount is zero", func(t *testing.T) {
		money, err := NewMoney(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, money.Amount())
	})

	t.Run("It should return an error when amount is negative", func(t *testing.T) {
		_, err := NewMoney(-50)
		assert.ErrorIs(t, err, apperr.ErrNegativeAmount)
	})

	t.Run("It should return an error when amount equals math.MaxInt", func(t *testing.T) {
		_, err := NewMoney(math.MaxInt)
		assert.ErrorIs(t, err, apperr.ErrAmountOverflow)
	})
}

func TestMoney_Amount(t *testing.T) {
	t.Run("It should return the amount when money is created with a valid amount", func(t *testing.T) {
		money, _ := NewMoney(250)
		assert.Equal(t, 250, money.Amount())
	})
}

func TestMoney_Sum(t *testing.T) {
	t.Run("It should return the sum when both amounts are valid", func(t *testing.T) {
		a, _ := NewMoney(100)
		b, _ := NewMoney(200)
		result, err := a.Sum(b)
		assert.NoError(t, err)
		assert.Equal(t, 300, result.Amount())
	})

	t.Run("It should return an error when result exceeds max int", func(t *testing.T) {
		a, _ := NewMoney(math.MaxInt - 1)
		b, _ := NewMoney(1)
		_, err := a.Sum(b)
		assert.ErrorIs(t, err, apperr.ErrAmountOverflow)
	})
}

func TestMoney_Sub(t *testing.T) {
	t.Run("It should return the difference when result is a valid positive amount", func(t *testing.T) {
		a, _ := NewMoney(300)
		b, _ := NewMoney(100)
		result, err := a.Sub(b)
		assert.NoError(t, err)
		assert.Equal(t, 200, result.Amount())
	})

	t.Run("It should return a valid zero Money when result is zero", func(t *testing.T) {
		a, _ := NewMoney(100)
		b, _ := NewMoney(100)
		result, err := a.Sub(b)
		assert.NoError(t, err)
		assert.Equal(t, 0, result.Amount())
	})

	t.Run("It should return an error when result is negative", func(t *testing.T) {
		a, _ := NewMoney(50)
		b, _ := NewMoney(100)
		_, err := a.Sub(b)
		assert.ErrorIs(t, err, apperr.ErrNegativeAmount)
	})
}

func TestMoney_Compare(t *testing.T) {
	t.Run("It should return zero when both amounts are equal", func(t *testing.T) {
		a, _ := NewMoney(100)
		b, _ := NewMoney(100)
		assert.Equal(t, 0, a.Compare(b))
	})

	t.Run("It should return a negative value when amount is less than other", func(t *testing.T) {
		a, _ := NewMoney(50)
		b, _ := NewMoney(100)
		assert.Less(t, a.Compare(b), 0)
	})

	t.Run("It should return a positive value when amount is greater than other", func(t *testing.T) {
		a, _ := NewMoney(200)
		b, _ := NewMoney(100)
		assert.Greater(t, a.Compare(b), 0)
	})
}
