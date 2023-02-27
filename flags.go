package main

import (
	"flag"
)

type config struct {
	listenHost string
	listenPort string
	role       string
	nick       string
}

func parseFlags() *config {
	c := &config{}

	flag.StringVar(&c.listenHost, "h", "0.0.0.0", "The bootstrap node host listen address\n")
	flag.StringVar(&c.listenPort, "p", "4001", "Node listen port")
	flag.StringVar(&c.role, "r", "user", "\n")
	flag.StringVar(&c.nick, "n", "", "\n")

	flag.Parse()
	return c
}
