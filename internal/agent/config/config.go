package config

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type Config struct {
	PullInterval   int64
	ReportInterval int64
	ReportAddress  string
}

func MustLoad() Config {
	var conf Config

	flag.Int64Var(&conf.PullInterval, "p", 2, "interval for pulling metrics in seconds")
	flag.Int64Var(&conf.ReportInterval, "r", 10, "interval for reporting in seconds")
	flag.StringVar(&conf.ReportAddress, "a", "localhost:8080", "http address for reporting")

	flag.Parse()

	if address := os.Getenv("ADDRESS"); address != "" {
		conf.ReportAddress = address
	}

	if val := os.Getenv("REPORT_INTERVAL"); val != "" {
		interval, err := strconv.Atoi(val)
		if err != nil {
			log.Fatal("cant parse REPORT_INTERVAL:", err)
		}

		conf.ReportInterval = int64(interval)
	}

	if val := os.Getenv("POLL_INTERVAL"); val != "" {
		interval, err := strconv.Atoi(val)
		if err != nil {
			log.Fatal("cant parse POLL_INTERVAL:", err)
		}

		conf.PullInterval = int64(interval)
	}

	return conf
}
