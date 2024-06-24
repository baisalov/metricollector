package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
	"log"
)

type Config struct {
	Address string `env:"ADDRESS"`
}

func MustLoad() Config {
	var conf Config

	flag.StringVar(&conf.Address, "a", "localhost:8080", "server running address")

	flag.Parse()

	err := env.Parse(&conf)
	if err != nil {
		log.Fatalf("Failed to load environments: %s", err.Error())
	}

	return conf
}
