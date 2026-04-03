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
	mySQLProjectorDAO "wallet/wallet-service/internal/projector/mysql_projector_dao"
	postgreSQLProjectorDAO "wallet/wallet-service/internal/projector/postgresql_projector_dao"
	"wallet/wallet-service/internal/query"
	"wallet/wallet-service/pkg/platform/subscriber"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load("../../.env")
	cfg := config.Load()

	time.Local = cfg.TZ

	settings, err := kurrentdb.ParseConnectionString(cfg.KurrentDBConnectionString)
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

	mySQLDB, err := sql.Open("mysql", cfg.MySQLConnectionString)
	if err != nil {
		slog.Error("failed to open mysql connection", slog.String("error", err.Error()))
		os.Exit(1)
	}
	mySQLDB.SetMaxOpenConns(cfg.MySQLMaxOpenConns)
	mySQLDB.SetMaxIdleConns(cfg.MySQLMaxIdleConns)
	mySQLDB.SetConnMaxLifetime(cfg.MySQLConnMaxLifetime)
	mySQLDB.SetConnMaxIdleTime(cfg.MySQLConnMaxIdleTime)

	if err := mySQLDB.Ping(); err != nil {
		slog.Error("failed to ping mysql", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer mySQLDB.Close()

	postgreSQLDB, err := sql.Open("postgres", cfg.PostgreSQLConnectionString)
	if err != nil {
		slog.Error("failed to open postgresql connection", slog.String("error", err.Error()))
		os.Exit(1)
	}
	postgreSQLDB.SetMaxOpenConns(cfg.PostgreSQLMaxOpenConns)
	postgreSQLDB.SetMaxIdleConns(cfg.PostgreSQLMaxIdleConns)
	postgreSQLDB.SetConnMaxLifetime(cfg.PostgreSQLConnMaxLifetime)
	postgreSQLDB.SetConnMaxIdleTime(cfg.PostgreSQLConnMaxIdleTime)

	if err := postgreSQLDB.Ping(); err != nil {
		slog.Error("failed to ping postgresql", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer postgreSQLDB.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := client.CreatePersistentSubscriptionToAll(
		ctx,
		cfg.WalletMySQLProjectionGroupName,
		kurrentdb.PersistentAllSubscriptionOptions{
			Filter: &kurrentdb.SubscriptionFilter{
				Type:     kurrentdb.EventFilterType,
				Prefixes: []string{"wallet:"},
			},
		},
	); err != nil && !subscriber.IsKurrentDBAlreadyExistsError(err) {
		slog.Error("failed to create mysql projection subscription", slog.String("group", cfg.WalletMySQLProjectionGroupName), slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := client.CreatePersistentSubscriptionToAll(
		ctx,
		cfg.WalletPostgreSQLProjectionGroupName,
		kurrentdb.PersistentAllSubscriptionOptions{
			Filter: &kurrentdb.SubscriptionFilter{
				Type:     kurrentdb.EventFilterType,
				Prefixes: []string{"wallet:"},
			},
		},
	); err != nil && !subscriber.IsKurrentDBAlreadyExistsError(err) {
		slog.Error("failed to create postgresql projection subscription", slog.String("group", cfg.WalletPostgreSQLProjectionGroupName), slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := client.CreatePersistentSubscriptionToAll(
		ctx,
		cfg.WalletCommandGroupName,
		kurrentdb.PersistentAllSubscriptionOptions{
			Filter: &kurrentdb.SubscriptionFilter{
				Type:     kurrentdb.EventFilterType,
				Prefixes: []string{"wallet:funds_transferred_event"},
			},
		},
	); err != nil && !subscriber.IsKurrentDBAlreadyExistsError(err) {
		slog.Error("failed to create command subscription", slog.String("group", cfg.WalletCommandGroupName), slog.String("error", err.Error()))
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

	walletConsumer := consumer.NewWalletKurrentDBConsumer(cfg.WalletCommandGroupName, client, walletESHandler)
	go walletConsumer.Start(ctx)

	mySQLProjectionDAO := mySQLProjectorDAO.NewWalletMySQLProjectionDAO(mySQLDB)
	mySQLProjectionConsumer := projector.NewWalletKurrentDBProjectorConsumer(cfg.WalletMySQLProjectionGroupName, client, mySQLProjectionDAO)
	go mySQLProjectionConsumer.Start(ctx)

	postgreSQLProjectionDAO := postgreSQLProjectorDAO.NewWalletPostgreSQLProjectionDAO(postgreSQLDB)
	postgreSQLProjectionConsumer := projector.NewWalletKurrentDBProjectorConsumer(cfg.WalletPostgreSQLProjectionGroupName, client, postgreSQLProjectionDAO)
	go postgreSQLProjectionConsumer.Start(ctx)

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
