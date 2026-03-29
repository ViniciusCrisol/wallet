package config

import "os"

type Config struct {
	MySQLConnectionString     string
	KurrentDBConnectionString string
	WalletProjectionGroupName string
	WalletCommandGroupName    string
	ServerAddress             string
}

func Load() Config {
	serverAddr := os.Getenv("SERVER_ADDRESS")
	if serverAddr == "" {
		serverAddr = ":8080"
	}
	return Config{
		MySQLConnectionString:     os.Getenv("MYSQL_CONNECTION_STRING"),
		KurrentDBConnectionString: os.Getenv("KURRENTDB_CONNECTION_STRING"),
		WalletProjectionGroupName: os.Getenv("WALLET_PROJECTION_GROUP_NAME"),
		WalletCommandGroupName:    os.Getenv("WALLET_COMMAND_GROUP_NAME"),
		ServerAddress:             serverAddr,
	}
}
