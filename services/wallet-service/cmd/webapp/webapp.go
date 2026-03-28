package main

import (
	"log"
	"net/http"

	"wallet/wallet-service/config"
	"wallet/wallet-service/internal/infrastructure/controller"
	"wallet/wallet-service/internal/infrastructure/persistence"

	"github.com/joho/godotenv"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

func main() {
	godotenv.Load()
	cfg := config.Load()

	settings, err := kurrentdb.ParseConnectionString(cfg.KurrentDBConnectionString)
	if err != nil {
		log.Fatalf("failed to parse kurrentdb connection string: %v", err)
	}
	db, err := kurrentdb.NewClient(settings)
	if err != nil {
		log.Fatalf("failed to connect to kurrentdb: %v", err)
	}
	defer db.Close()

	dao := persistence.NewWalletKurrentDBDAO(db)
	ctrl := controller.NewWalletController(dao)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /wallets", ctrl.Create)
	mux.HandleFunc("POST /wallets/{id}/transfer", ctrl.TransferFunds)
	mux.HandleFunc("POST /wallets/{id}/mock-transfer", ctrl.MockTransfer)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
