package query

import (
	"database/sql"
	"net/http"
	"os"
	"testing"
	"time"

	"wallet/wallet-service/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var db *sql.DB

func TestMain(m *testing.M) {
	godotenv.Load("../../.env.test")
	cfg := config.Load()

	time.Local = cfg.TZ

	mysql, err := sql.Open("mysql", cfg.MySQLConnectionString)
	if err != nil {
		panic(err)
	}
	mysql.SetMaxOpenConns(cfg.MySQLMaxOpenConns)
	mysql.SetMaxIdleConns(cfg.MySQLMaxIdleConns)
	mysql.SetConnMaxLifetime(cfg.MySQLConnMaxLifetime)
	mysql.SetConnMaxIdleTime(cfg.MySQLConnMaxIdleTime)

	if err = mysql.Ping(); err != nil {
		panic(err)
	}
	db = mysql

	os.Exit(m.Run())
}

func createTestWallet(t *testing.T, walletID string, holderID string) {
	t.Helper()

	_, err := db.Exec(
		"INSERT INTO wallet_projections (wallet_id, holder_id, balance_in_cents, created_at, updated_at) VALUES (?, ?, 0, ?, ?)",
		walletID,
		holderID,
		time.Now(),
		time.Now(),
	)
	assert.NoError(t, err)
}

func createTestTransfer(
	t *testing.T,
	walletID string,
	transferID string,
	counterpartWalletID string,
	direction string,
	amountInCents int,
	transferredAt time.Time,
) {
	t.Helper()

	_, err := db.Exec(
		`INSERT INTO transfer_projections (wallet_id, transfer_id, counterpart_wallet_id, direction, amount_in_cents, transferred_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		walletID,
		transferID,
		counterpartWalletID,
		direction,
		amountInCents,
		transferredAt,
	)
	assert.NoError(t, err)
}

func newFindByIDMux(t *testing.T, controller *WalletQueryController) *http.ServeMux {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /wallets/{id}", controller.FindByID)
	return mux
}

func newFindByHolderIDMux(t *testing.T, controller *WalletQueryController) *http.ServeMux {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /wallets", controller.FindByHolderID)
	return mux
}

func newFindTransfersMux(t *testing.T, controller *WalletQueryController) *http.ServeMux {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /wallets/{id}/transfers", controller.FindTransfersByWalletID)
	return mux
}
