package projector

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"wallet/wallet-service/config"
	"wallet/wallet-service/pkg/integrationevent"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
)

var (
	db              *sql.DB
	kurrentDBClient *kurrentdb.Client
)

func TestMain(m *testing.M) {
	godotenv.Load("../../.env.test")

	d, err := sql.Open("mysql", os.Getenv("MYSQL_CONNECTION_STRING"))
	if err != nil {
		log.Fatal(err)
	}
	if err = d.Ping(); err != nil {
		log.Fatal(err)
	}
	db = d

	settings, err := kurrentdb.ParseConnectionString(config.Load().KurrentDBConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	kurrentDBClient, err = kurrentdb.NewClient(settings)
	if err != nil {
		log.Fatal(err)
	}
	defer kurrentDBClient.Close()

	os.Exit(m.Run())
}

func createTestWallet(
	t *testing.T,
	walletID string,
	holderID string,
	walletMySQLProjectionDAO *WalletMySQLProjectionDAO,
) {
	event := integrationevent.WalletCreatedEvent{
		WalletID:  walletID,
		HolderID:  holderID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	assert.NoError(t, walletMySQLProjectionDAO.CreateWallet(context.Background(), event))
}

func getTestWalletBalance(t *testing.T, walletID string) int {
	var balance int
	assert.NoError(t, db.QueryRow("SELECT balance_in_cents FROM wallet_projections WHERE wallet_id = ?", walletID).Scan(&balance))
	return balance
}
