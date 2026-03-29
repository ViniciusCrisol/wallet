package query

import (
	"database/sql"
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
