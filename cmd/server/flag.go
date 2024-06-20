package main

import (
	"flag"
	"os"
)

var runningAddress string

func init() {
	flag.StringVar(&runningAddress, "a", "localhost:8080", "server running address")

	flag.Parse()

	if address := os.Getenv("ADDRESS"); address != "" {
		runningAddress = address
	}
}
