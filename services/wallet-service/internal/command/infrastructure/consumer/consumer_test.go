package consumer

import (
	"log"
	"os"
	"testing"

	"wallet/wallet-service/config"

	"github.com/joho/godotenv"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

var kurrentDBClient *kurrentdb.Client

func TestMain(m *testing.M) {
	godotenv.Load("../../../../.env.test")

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
