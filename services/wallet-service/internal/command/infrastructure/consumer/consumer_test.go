package consumer

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"wallet/wallet-service/config"
	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/internal/command/infrastructure/persistence"
	"wallet/wallet-service/pkg/domain/valueobject"

	"github.com/joho/godotenv"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/require"
)

var client *kurrentdb.Client

func TestMain(m *testing.M) {
	godotenv.Load("../../../../.env.test")

	settings, err := kurrentdb.ParseConnectionString(config.Load().KurrentDBConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	client, err = kurrentdb.NewClient(settings)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	os.Exit(m.Run())
}

func createTestWalletWithBalance(t *testing.T, esHandler *persistence.WalletKurrentDBESHandler, balance int) domain.Wallet {
	t.Helper()

	wallet := domain.NewWallet(domain.CreateWalletCommand{
		WalletID:  valueobject.GenerateID(),
		HolderID:  valueobject.GenerateID(),
		Timestamp: time.Now(),
	})
	if balance > 0 {
		amount, err := valueobject.NewMoney(balance)
		require.NoError(t, err)
		require.NoError(t, wallet.ReceiveFundsTransfer(domain.ReceiveFundsTransferCommand{
			Amount:       amount,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		}))
	}
	require.NoError(t, esHandler.Save(context.Background(), wallet))
	return wallet
}
