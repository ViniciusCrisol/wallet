package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wallet/wallet-service/config"
	"wallet/wallet-service/internal/command/infrastructure/consumer"
	"wallet/wallet-service/internal/command/infrastructure/controller"
	"wallet/wallet-service/internal/command/infrastructure/persistence"
	"wallet/wallet-service/internal/projection"
	"wallet/wallet-service/internal/query"
	"wallet/wallet-service/pkg/eventsourcing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

func main() {
	godotenv.Load()
	cfg := config.Load()

	kurrenDBSettings, err := kurrentdb.ParseConnectionString(cfg.KurrentDBConnectionString)
	if err != nil {
		slog.Error("failed to parse kurrentdb connection string", slog.String("error", err.Error()))
		os.Exit(1)
	}
	kurrenDBSettings.Logger = kurrentdb.NoopLogging()
	kurrentDBClient, err := kurrentdb.NewClient(kurrenDBSettings)
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

	if err := kurrentDBClient.CreatePersistentSubscriptionToAll(
		ctx,
		cfg.WalletProjectionGroupName,
		kurrentdb.PersistentAllSubscriptionOptions{
			Filter: &kurrentdb.SubscriptionFilter{
				Type:     kurrentdb.EventFilterType,
				Prefixes: []string{"wallet:"},
			},
		},
	); err != nil && !eventsourcing.IsKurrentDBAlreadyExistsError(err) {
		slog.Error("failed to create projection subscription", slog.String("group", cfg.WalletProjectionGroupName), slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := kurrentDBClient.CreatePersistentSubscriptionToAll(
		ctx,
		cfg.WalletCommandGroupName,
		kurrentdb.PersistentAllSubscriptionOptions{
			Filter: &kurrentdb.SubscriptionFilter{
				Type:     kurrentdb.EventFilterType,
				Prefixes: []string{"wallet:funds_transferred_event"},
			},
		},
	); err != nil && !eventsourcing.IsKurrentDBAlreadyExistsError(err) {
		slog.Error("failed to create command subscription", slog.String("group", cfg.WalletCommandGroupName), slog.String("error", err.Error()))
		os.Exit(1)
	}

	walletESHandler := persistence.NewWalletKurrentDBESHandler(kurrentDBClient)
	walletCommandController := controller.NewWalletCommandController(walletESHandler)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /wallets", walletCommandController.Create)
	mux.HandleFunc("POST /wallets/{id}/transfer", walletCommandController.TransferFunds)
	mux.HandleFunc("POST /wallets/{id}/mock-transfer", walletCommandController.MockTransfer)

	walletQueryController := query.NewWalletQueryController(mySQLDB)
	mux.HandleFunc("GET /wallets/{id}", walletQueryController.FindByID)
	mux.HandleFunc("GET /wallets", walletQueryController.FindByHolderID)

	walletConsumer := consumer.NewWalletKurrentDBConsumer(cfg.WalletCommandGroupName, kurrentDBClient, walletESHandler)
	go walletConsumer.Start(ctx)

	projectionDAO := projection.NewWalletMySQLProjectionDAO(mySQLDB)
	projectionConsumer := projection.NewWalletKurrentDBProjectionConsumer(cfg.WalletProjectionGroupName, kurrentDBClient, projectionDAO)
	go projectionConsumer.Start(ctx)

	server := &http.Server{Addr: cfg.ServerAddress, Handler: mux}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()
	slog.Info("http server starting", slog.String("addr", cfg.ServerAddress))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("http server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
