package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
	"log"
)

type Config struct {
	PullInterval   int64  `env:"POLL_INTERVAL"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	ReportAddress  string `env:"ADDRESS"`
	HashKey        string `env:"KEY"`
	ReteLimit      int    `env:"RATE_LIMIT"`
}

func MustLoad() Config {
	var conf Config

	flag.Int64Var(&conf.PullInterval, "p", 2, "interval for pulling metrics in seconds")
	flag.Int64Var(&conf.ReportInterval, "r", 10, "interval for reporting in seconds")
	flag.StringVar(&conf.ReportAddress, "a", "localhost:8080", "http address for reporting")
	flag.StringVar(&conf.HashKey, "k", "", "key for sign body hash")
	flag.IntVar(&conf.ReteLimit, "l", 10, "parallel senders limit")

	flag.Parse()

	err := env.Parse(&conf)
	if err != nil {
		log.Fatalf("Failed to load environments: %s", err.Error())
	}

	return conf
}
