package domain

import (
	"testing"
	"time"

	valueObject "wallet/wallet-service/pkg/domain/value_object"

	"github.com/stretchr/testify/assert"
)

func newMoney(t *testing.T, amountInCents int) valueObject.Money {
	t.Helper()

	money, err := valueObject.NewMoney(amountInCents)
	assert.NoError(t, err)
	return money
}

func newTestWallet(t *testing.T) Wallet {
	t.Helper()

	return NewWallet(CreateWalletCommand{
		WalletID:  valueObject.GenerateID(),
		HolderID:  valueObject.GenerateID(),
		Timestamp: time.Now(),
	})
}

func newTestWalletWithBalance(t *testing.T, amountInCents int) Wallet {
	t.Helper()

	wallet := newTestWallet(t)
	err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
		Amount:       newMoney(t, amountInCents),
		TransferID:   valueObject.GenerateID(),
		FromWalletID: valueObject.GenerateID(),
		Timestamp:    time.Now(),
	})
	assert.NoError(t, err)
	wallet.Commit()
	return wallet
}
