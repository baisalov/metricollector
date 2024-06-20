package main

import (
	"flag"
	"log"
	"os"
	"strconv"
)

var (
	pullInterval   int64
	reportInterval int64
	reportAddress  string
)

func init() {
	flag.Int64Var(&pullInterval, "p", 2, "interval for pulling metrics in seconds")
	flag.Int64Var(&reportInterval, "r", 10, "interval for reporting in seconds")
	flag.StringVar(&reportAddress, "a", "localhost:8080", "http address for reporting")

	flag.Parse()

	if address := os.Getenv("ADDRESS"); address != "" {
		reportAddress = address
	}

	if val := os.Getenv("REPORT_INTERVAL"); val != "" {
		interval, err := strconv.Atoi(val)
		if err != nil {
			log.Fatal("cant parse REPORT_INTERVAL:", err)
		}

		reportInterval = int64(interval)
	}

	if val := os.Getenv("POLL_INTERVAL"); val != "" {
		interval, err := strconv.Atoi(val)
		if err != nil {
			log.Fatal("cant parse POLL_INTERVAL:", err)
		}

		pullInterval = int64(interval)
	}
}
