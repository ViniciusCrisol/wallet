package projector

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"wallet/wallet-service/config"
	"wallet/wallet-service/pkg/platform/integrationevent"
	"wallet/wallet-service/pkg/platform/uuid"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
)

var (
	db     *sql.DB
	client *kurrentdb.Client
)

type transferProjectionRow struct {
	WalletID            string
	TransferID          string
	CounterpartWalletID string
	Direction           string
	AmountInCents       int
	TransferredAt       time.Time
}

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
	settings.Logger = kurrentdb.NoopLogging()

	client, err = kurrentdb.NewClient(settings)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	group := "test-group-" + uuid.NewUUID()
	client.CreatePersistentSubscriptionToAll(
		ctx,
		group,
		kurrentdb.PersistentAllSubscriptionOptions{},
	)
	go NewWalletKurrentDBProjectorConsumer(
		group,
		client,
		NewWalletMySQLProjectionDAO(db),
	).Start(ctx)

	os.Exit(m.Run())
}

func createTestWallet(
	t *testing.T,
	walletID string,
	holderID string,
	walletMySQLProjectionDAO *WalletMySQLProjectionDAO,
) {
	t.Helper()

	event := integrationevent.WalletCreatedEvent{
		WalletID:  walletID,
		HolderID:  holderID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	assert.NoError(t, walletMySQLProjectionDAO.CreateWallet(context.Background(), event))
}

func getTestWalletBalance(t *testing.T, walletID string) int {
	t.Helper()

	var balance int
	assert.NoError(t, db.QueryRow("SELECT balance_in_cents FROM wallet_projections WHERE wallet_id = ?", walletID).Scan(&balance))
	return balance
}

func getTestTransferProjection(
	t *testing.T,
	walletID string,
	transferID string,
) (transferProjectionRow, bool) {
	t.Helper()

	var row transferProjectionRow
	err := db.QueryRow(
		`
			SELECT
				wallet_id, transfer_id, counterpart_wallet_id, direction, amount_in_cents, transferred_at
			FROM
				transfer_projections
			WHERE
				wallet_id = ? AND transfer_id = ?
		`, walletID, transferID,
	).Scan(
		&row.WalletID,
		&row.TransferID,
		&row.CounterpartWalletID,
		&row.Direction,
		&row.AmountInCents,
		&row.TransferredAt,
	)
	if err != nil {
		return row, false
	}
	return row, true
}
