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
	"wallet/wallet-service/internal/projector"
	"wallet/wallet-service/internal/query"
	"wallet/wallet-service/pkg/platform/subscriber"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

func main() {
	godotenv.Load("../../.env")
	config := config.Load()

	settings, err := kurrentdb.ParseConnectionString(config.KurrentDBConnectionString)
	if err != nil {
		slog.Error("failed to parse kurrentdb connection string", slog.String("error", err.Error()))
		os.Exit(1)
	}
	settings.Logger = kurrentdb.NoopLogging()

	client, err := kurrentdb.NewClient(settings)
	if err != nil {
		slog.Error("failed to connect to kurrentdb", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer client.Close()

	mySQLDB, err := sql.Open("mysql", config.MySQLConnectionString)
	if err != nil {
		slog.Error("failed to open mysql connection", slog.String("error", err.Error()))
		os.Exit(1)
	}
	mySQLDB.SetMaxOpenConns(config.MySQLMaxOpenConns)
	mySQLDB.SetMaxIdleConns(config.MySQLMaxIdleConns)
	mySQLDB.SetConnMaxLifetime(config.MySQLConnMaxLifetime)
	mySQLDB.SetConnMaxIdleTime(config.MySQLConnMaxIdleTime)

	if err := mySQLDB.Ping(); err != nil {
		slog.Error("failed to ping mysql", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer mySQLDB.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := client.CreatePersistentSubscriptionToAll(
		ctx,
		config.WalletProjectionGroupName,
		kurrentdb.PersistentAllSubscriptionOptions{
			Filter: &kurrentdb.SubscriptionFilter{
				Type:     kurrentdb.EventFilterType,
				Prefixes: []string{"wallet:"},
			},
		},
	); err != nil && !subscriber.IsKurrentDBAlreadyExistsError(err) {
		slog.Error("failed to create projection subscription", slog.String("group", config.WalletProjectionGroupName), slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := client.CreatePersistentSubscriptionToAll(
		ctx,
		config.WalletCommandGroupName,
		kurrentdb.PersistentAllSubscriptionOptions{
			Filter: &kurrentdb.SubscriptionFilter{
				Type:     kurrentdb.EventFilterType,
				Prefixes: []string{"wallet:funds_transferred_event"},
			},
		},
	); err != nil && !subscriber.IsKurrentDBAlreadyExistsError(err) {
		slog.Error("failed to create command subscription", slog.String("group", config.WalletCommandGroupName), slog.String("error", err.Error()))
		os.Exit(1)
	}

	walletESHandler := persistence.NewWalletKurrentDBESHandler(client)
	walletCommandController := controller.NewWalletCommandController(walletESHandler)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /wallets", walletCommandController.Create)
	mux.HandleFunc("POST /wallets/{id}/transfer", walletCommandController.TransferFunds)
	mux.HandleFunc("POST /wallets/{id}/mock-transfer", walletCommandController.MockTransfer)

	walletQueryController := query.NewWalletQueryController(mySQLDB)
	mux.HandleFunc("GET /wallets", walletQueryController.FindByHolderID)
	mux.HandleFunc("GET /wallets/{id}", walletQueryController.FindByID)
	mux.HandleFunc("GET /wallets/{id}/transfers", walletQueryController.FindTransfersByWalletID)

	walletConsumer := consumer.NewWalletKurrentDBConsumer(config.WalletCommandGroupName, client, walletESHandler)
	go walletConsumer.Start(ctx)

	projectionDAO := projector.NewWalletMySQLProjectionDAO(mySQLDB)
	projectionConsumer := projector.NewWalletKurrentDBProjectorConsumer(config.WalletProjectionGroupName, client, projectionDAO)
	go projectionConsumer.Start(ctx)

	server := &http.Server{Addr: config.ServerAddress, Handler: mux}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()
	slog.Info("http server starting", slog.String("addr", config.ServerAddress))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("http server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
