package config

import "os"

type Config struct {
	ServerAddress             string
	MySQLConnectionString     string
	KurrentDBConnectionString string
	WalletCommandGroupName    string
	WalletProjectionGroupName string
}

func Load() Config {
	return Config{
		ServerAddress:             os.Getenv("SERVER_ADDRESS"),
		MySQLConnectionString:     os.Getenv("MYSQL_CONNECTION_STRING"),
		KurrentDBConnectionString: os.Getenv("KURRENTDB_CONNECTION_STRING"),
		WalletCommandGroupName:    os.Getenv("WALLET_COMMAND_GROUP_NAME"),
		WalletProjectionGroupName: os.Getenv("WALLET_PROJECTION_GROUP_NAME"),
	}
}
