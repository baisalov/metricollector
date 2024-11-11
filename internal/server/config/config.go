package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
	"log"
)

type Config struct {
	Address       string `env:"ADDRESS"`
	StoragePath   string `env:"FILE_STORAGE_PATH" envDefault:"storage.txt"`
	StoreInterval int64  `env:"STORE_INTERVAL" envDefault:"300"`
	Restore       bool   `env:"RESTORE" envDefault:"true"`
	DatabaseDsn   string `env:"DATABASE_DSN"`
	HashKey       string `env:"KEY"`
}

func MustLoad() Config {
	var conf Config

	flag.StringVar(&conf.Address, "a", "localhost:8080", "server running address")

	flag.StringVar(&conf.StoragePath, "f", "storage.txt", "file storage path")
	flag.Int64Var(&conf.StoreInterval, "i", 300, "flush to file storage interval on seconds (0 - sync store)")
	flag.BoolVar(&conf.Restore, "r", true, "restore storage from file when running")
	flag.StringVar(&conf.DatabaseDsn, "d", "", "dsn for connection to database")

	err := env.Parse(&conf)
	if err != nil {
		log.Fatalf("Failed to load environments: %s", err.Error())
	}

	flag.Parse()

	return conf
}
