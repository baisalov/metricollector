package main

import "flag"

var runningAddress string

func init() {
	flag.StringVar(&runningAddress, "a", "localhost:8080", "server running address")

	flag.Parse()
}
