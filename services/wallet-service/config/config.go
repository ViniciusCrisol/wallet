package config

import "os"

type Config struct {
	KurrentDBConnectionString string
}

func Load() Config {
	return Config{
		KurrentDBConnectionString: os.Getenv("KURRENTDB_CONNECTION_STRING"),
	}
}
