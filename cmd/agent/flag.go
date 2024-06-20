package main

import (
	"flag"
	"time"
)

var (
	pullInterval   time.Duration
	reportInterval time.Duration
	reportAddress  string
)

func init() {
	flag.DurationVar(&pullInterval, "p", 2*time.Second, "interval for receiving metrics")
	flag.DurationVar(&reportInterval, "r", 10*time.Second, "interval for reporting")
	flag.StringVar(&reportAddress, "a", "localhost:8080", "http address for reporting")

	flag.Parse()
}
