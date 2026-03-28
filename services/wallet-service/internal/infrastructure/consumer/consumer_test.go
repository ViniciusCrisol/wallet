package consumer

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	godotenv.Load("../../../.env.test")

	var err error
	testDB, err = sql.Open("mysql", os.Getenv("MYSQL_CONNECTION_STRING"))
	if err != nil {
		panic(err)
	}
	if err = testDB.Ping(); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}
