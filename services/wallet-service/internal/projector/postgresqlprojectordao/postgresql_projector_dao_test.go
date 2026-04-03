package postgresqlprojectordao

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"wallet/wallet-service/pkg/platform/integrationevent"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var postgreSQLDB *sql.DB

type transferProjectionRow struct {
	WalletID            string
	TransferID          string
	CounterpartWalletID string
	Direction           string
	AmountInCents       int
	TransferredAt       time.Time
}

func TestMain(m *testing.M) {
	godotenv.Load("../../../.env.test")

	pg, err := sql.Open("postgres", os.Getenv("POSTGRESQL_CONNECTION_STRING"))
	if err != nil {
		log.Fatal(err)
	}
	if err = pg.Ping(); err != nil {
		log.Fatal(err)
	}
	postgreSQLDB = pg

	os.Exit(m.Run())
}

func createTestWallet(
	t *testing.T,
	walletID string,
	holderID string,
	walletPostgreSQLProjectionDAO *WalletPostgreSQLProjectionDAO,
) {
	t.Helper()

	event := integrationevent.WalletCreatedEvent{
		WalletID:  walletID,
		HolderID:  holderID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	assert.NoError(t, walletPostgreSQLProjectionDAO.CreateWallet(context.Background(), event))
}

func getTestWalletBalance(t *testing.T, walletID string) int {
	t.Helper()

	var balance int
	assert.NoError(t, postgreSQLDB.QueryRow("SELECT balance_in_cents FROM wallet_projections WHERE wallet_id = $1", walletID).Scan(&balance))
	return balance
}

func getTestTransferProjection(
	t *testing.T,
	walletID string,
	transferID string,
) (transferProjectionRow, bool) {
	t.Helper()

	var row transferProjectionRow
	err := postgreSQLDB.QueryRow(
		`
			SELECT
				wallet_id, transfer_id, counterpart_wallet_id, direction, amount_in_cents, transferred_at
			FROM
				transfer_projections
			WHERE
				wallet_id = $1 AND transfer_id = $2
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
