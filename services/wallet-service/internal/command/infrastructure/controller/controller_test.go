package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"wallet/wallet-service/config"
	"wallet/wallet-service/internal/command/infrastructure/persistence"

	"github.com/joho/godotenv"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

func TestMain(m *testing.M) {
	godotenv.Load("../../../../.env.test")

	time.Local = config.Load().TZ

	os.Exit(m.Run())
}

func newTestController(t *testing.T) *WalletCommandController {
	t.Helper()

	settings, err := kurrentdb.ParseConnectionString(config.Load().KurrentDBConnectionString)
	if err != nil {
		t.Fatalf("failed to parse connection string: %v", err)
	}
	settings.Logger = kurrentdb.NoopLogging()

	db, err := kurrentdb.NewClient(settings)
	if err != nil {
		t.Fatalf("failed to create kurrentdb client: %v", err)
	}
	return &WalletCommandController{
		esHandler: persistence.NewWalletKurrentDBESHandler(db),
	}
}

func marshalBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()

	j, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}
	return bytes.NewBuffer(j)
}

func newCreateMux(controller *WalletCommandController) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /wallets", controller.Create)
	return mux
}

func newTransferMux(controller *WalletCommandController) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /wallets/{id}/transfer", controller.TransferFunds)
	return mux
}

func newMockTransferMux(controller *WalletCommandController) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /wallets/{id}/mock-transfer", controller.MockTransfer)
	return mux
}
