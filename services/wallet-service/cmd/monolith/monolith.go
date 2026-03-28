package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"wallet/wallet-service/config"
	"wallet/wallet-service/internal/commandside/infrastructure/controller"
	"wallet/wallet-service/internal/commandside/infrastructure/persistence"
	"wallet/wallet-service/internal/projectionbuilder"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

func main() {
	godotenv.Load()
	cfg := config.Load()

	settings, err := kurrentdb.ParseConnectionString(cfg.KurrentDBConnectionString)
	if err != nil {
		slog.Error("failed to parse kurrentdb connection string", slog.String("error", err.Error()))
		os.Exit(1)
	}
	kurrentDBClient, err := kurrentdb.NewClient(settings)
	if err != nil {
		slog.Error("failed to connect to kurrentdb", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer kurrentDBClient.Close()

	mySQLDB, err := sql.Open("mysql", cfg.MySQLConnectionString)
	if err != nil {
		slog.Error("failed to open mysql connection", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := mySQLDB.Ping(); err != nil {
		slog.Error("failed to ping mysql", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer mySQLDB.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	kurrentDBClient.CreatePersistentSubscriptionToAll(
		ctx,
		cfg.WalletProjectionGroupName,
		kurrentdb.PersistentAllSubscriptionOptions{},
	)

	walletController := controller.NewWalletController(persistence.NewWalletKurrentDBESHandler(kurrentDBClient))
	mux := http.NewServeMux()
	mux.HandleFunc("POST /wallets", walletController.Create)
	mux.HandleFunc("POST /wallets/{id}/transfer", walletController.TransferFunds)
	mux.HandleFunc("POST /wallets/{id}/mock-transfer", walletController.MockTransfer)

	projectionDAO := projectionbuilder.NewWalletMySQLProjectionDAO(mySQLDB)
	projectionConsumer := projectionbuilder.NewWalletKurrentDBProjectionConsumer(kurrentDBClient, projectionDAO)
	go projectionConsumer.Start(ctx)

	slog.Info("http server starting", slog.String("addr", ":8080"))
	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("http server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
