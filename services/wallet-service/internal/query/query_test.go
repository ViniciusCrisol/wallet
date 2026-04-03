package query

import (
	"database/sql"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var db *sql.DB

func TestMain(m *testing.M) {
	godotenv.Load("../../.env.test")

	d, err := sql.Open("mysql", os.Getenv("MYSQL_CONNECTION_STRING"))
	if err != nil {
		panic(err)
	}
	if err = d.Ping(); err != nil {
		panic(err)
	}
	db = d

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
