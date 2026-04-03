package persistence

import (
	"os"
	"testing"
	"time"

	"wallet/wallet-service/config"
	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/pkg/domain/valueobject"

	"github.com/joho/godotenv"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

func TestMain(m *testing.M) {
	godotenv.Load("../../../../.env.test")

	os.Exit(m.Run())
}

func newWallet(t *testing.T) domain.Wallet {
	t.Helper()

	return domain.NewWallet(domain.CreateWalletCommand{
		WalletID:  valueobject.GenerateID(),
		HolderID:  valueobject.GenerateID(),
		Timestamp: time.Now(),
	})
}

func newTestESHandler(t *testing.T) *WalletKurrentDBESHandler {
	t.Helper()

	settings, err := kurrentdb.ParseConnectionString(config.Load().KurrentDBConnectionString)
	if err != nil {
		t.Fatalf("failed to parse connection string: %v", err)
	}
	db, err := kurrentdb.NewClient(settings)
	if err != nil {
		t.Fatalf("failed to create kurrentdb client: %v", err)
	}
	return NewWalletKurrentDBESHandler(db)
}
