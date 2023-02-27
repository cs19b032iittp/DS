package main

import "log"

func main() {
	log.Println("Application Started,")

	cfg := parseFlags()

	switch cfg.role {
	case "tracker":
		CreateTracker(cfg)
	case "user":
		CreateUser(cfg)
	}

	select {}
}

func Log(FuncCall string, PeerName string) {
	log.Printf("\x1b[36m Got %s RPC from %s \x1b[0m\n", FuncCall, PeerName)
}
