package config

import "os"

type Config struct {
	MySQLConnectionString        string
	KurrentDBConnectionString    string
	WalletProjectionSubscription string
}

func Load() Config {
	return Config{
		MySQLConnectionString:        os.Getenv("MYSQL_CONNECTION_STRING"),
		KurrentDBConnectionString:    os.Getenv("KURRENTDB_CONNECTION_STRING"),
		WalletProjectionSubscription: os.Getenv("WALLET_PROJECTION_SUBSCRIPTION"),
	}
}
