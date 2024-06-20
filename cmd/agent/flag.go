package main

import (
	"flag"
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
}
