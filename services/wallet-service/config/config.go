package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerAddress                       string
	MySQLConnectionString               string
	MySQLMaxOpenConns                   int
	MySQLMaxIdleConns                   int
	MySQLConnMaxLifetime                time.Duration
	MySQLConnMaxIdleTime                time.Duration
	PostgreSQLConnectionString          string
	PostgreSQLMaxOpenConns              int
	PostgreSQLMaxIdleConns              int
	PostgreSQLConnMaxLifetime           time.Duration
	PostgreSQLConnMaxIdleTime           time.Duration
	KurrentDBConnectionString           string
	WalletCommandGroupName              string
	WalletMySQLProjectionGroupName      string
	WalletPostgreSQLProjectionGroupName string
}

func Load() Config {
	return Config{
		ServerAddress:                       os.Getenv("SERVER_ADDRESS"),
		MySQLConnectionString:               os.Getenv("MYSQL_CONNECTION_STRING"),
		MySQLMaxOpenConns:                   envInt("MYSQL_MAX_OPEN_CONNS"),
		MySQLMaxIdleConns:                   envInt("MYSQL_MAX_IDLE_CONNS"),
		MySQLConnMaxLifetime:                envDuration("MYSQL_CONN_MAX_LIFETIME"),
		MySQLConnMaxIdleTime:                envDuration("MYSQL_CONN_MAX_IDLE_TIME"),
		PostgreSQLConnectionString:          os.Getenv("POSTGRESQL_CONNECTION_STRING"),
		PostgreSQLMaxOpenConns:              envInt("POSTGRESQL_MAX_OPEN_CONNS"),
		PostgreSQLMaxIdleConns:              envInt("POSTGRESQL_MAX_IDLE_CONNS"),
		PostgreSQLConnMaxLifetime:           envDuration("POSTGRESQL_CONN_MAX_LIFETIME"),
		PostgreSQLConnMaxIdleTime:           envDuration("POSTGRESQL_CONN_MAX_IDLE_TIME"),
		KurrentDBConnectionString:           os.Getenv("KURRENTDB_CONNECTION_STRING"),
		WalletCommandGroupName:              os.Getenv("WALLET_COMMAND_GROUP_NAME"),
		WalletMySQLProjectionGroupName:      os.Getenv("WALLET_MYSQL_PROJECTION_GROUP_NAME"),
		WalletPostgreSQLProjectionGroupName: os.Getenv("WALLET_POSTGRESQL_PROJECTION_GROUP_NAME"),
	}
}

func envInt(key string) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return 0
}

func envDuration(key string) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return 0
}
