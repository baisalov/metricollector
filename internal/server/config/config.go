package config

import (
	"flag"
	"os"
)

type Config struct {
	Address string
}

func MustLoad() Config {
	var conf Config

	flag.StringVar(&conf.Address, "a", "localhost:8080", "server running address")

	flag.Parse()

	if address := os.Getenv("ADDRESS"); address != "" {
		conf.Address = address
	}

	return conf
}
