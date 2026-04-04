package mysql_projector_dao

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"wallet/wallet-service/config"
	integrationEvent "wallet/wallet-service/pkg/platform/integration_event"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var db *sql.DB

type transferProjectionRow struct {
	WalletID            string
	TransferID          string
	CounterpartWalletID string
	Direction           string
	Category            string
	AmountInCents       int
	TransferredAt       time.Time
}

func TestMain(m *testing.M) {
	godotenv.Load("../../../.env.test")
	cfg := config.Load()

	time.Local = cfg.TZ

	mysql, err := sql.Open("mysql", cfg.MySQLConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	mysql.SetMaxOpenConns(cfg.MySQLMaxOpenConns)
	mysql.SetMaxIdleConns(cfg.MySQLMaxIdleConns)
	mysql.SetConnMaxLifetime(cfg.MySQLConnMaxLifetime)
	mysql.SetConnMaxIdleTime(cfg.MySQLConnMaxIdleTime)

	if err = mysql.Ping(); err != nil {
		log.Fatal(err)
	}
	db = mysql

	os.Exit(m.Run())
}

func createTestWallet(
	t *testing.T,
	walletID string,
	holderID string,
	walletMySQLProjectionDAO *WalletMySQLProjectionDAO,
) {
	t.Helper()

	event := integrationEvent.WalletCreatedEvent{
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
				wallet_id, transfer_id, counterpart_wallet_id, direction, category, amount_in_cents, transferred_at
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
		&row.Category,
		&row.AmountInCents,
		&row.TransferredAt,
	)
	if err != nil {
		return row, false
	}
	return row, true
}
