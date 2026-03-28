package main

import (
	"context"
	"database/sql"
	"log"
	"os/signal"
	"syscall"

	"wallet/wallet-service/config"
	"wallet/wallet-service/internal/infrastructure/consumer"
	"wallet/wallet-service/internal/infrastructure/persistence"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

func main() {
	godotenv.Load()
	cfg := config.Load()

	db, err := sql.Open("mysql", cfg.MySQLConnectionString)
	if err != nil {
		log.Fatalf("failed to open mysql connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping mysql: %v", err)
	}

	settings, err := kurrentdb.ParseConnectionString(cfg.KurrentDBConnectionString)
	if err != nil {
		log.Fatalf("failed to parse kurrentdb connection string: %v", err)
	}
	client, err := kurrentdb.NewClient(settings)
	if err != nil {
		log.Fatalf("failed to connect to kurrentdb: %v", err)
	}
	defer client.Close()

	client.CreatePersistentSubscriptionToAll(
		context.TODO(),
		cfg.WalletProjectionSubscription,
		kurrentdb.PersistentAllSubscriptionOptions{},
	)

	dao := persistence.NewWalletMysqlDAO(db)
	projConsumer := consumer.NewProjectionConsumer(client, dao)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	projConsumer.Start(ctx)
}
