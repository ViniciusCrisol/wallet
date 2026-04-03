package domain

import (
	"testing"
	"time"

	"wallet/wallet-service/pkg/valueobject"

	"github.com/stretchr/testify/assert"
)

func newMoney(t *testing.T, amountInCents int) valueobject.Money {
	t.Helper()

	money, err := valueobject.NewMoney(amountInCents)
	assert.NoError(t, err)
	return money
}

func newTestWallet(t *testing.T) Wallet {
	t.Helper()

	return NewWallet(CreateWalletCommand{
		WalletID:  valueobject.GenerateID(),
		HolderID:  valueobject.GenerateID(),
		Timestamp: time.Now(),
	})
}

func newTestWalletWithBalance(t *testing.T, amountInCents int) Wallet {
	t.Helper()

	wallet := newTestWallet(t)
	err := wallet.ReceiveFundsTransfer(ReceiveFundsTransferCommand{
		Amount:       newMoney(t, amountInCents),
		TransferID:   valueobject.GenerateID(),
		FromWalletID: valueobject.GenerateID(),
		Timestamp:    time.Now(),
	})
	assert.NoError(t, err)
	wallet.Commit()
	return wallet
}
