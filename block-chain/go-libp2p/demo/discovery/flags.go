package main

import (
	"flag"
)

type config struct {
	RendezvousString string
	id               string
	ProtocolID       string
	listenHost       string
	bootstrap        string
	listenPort       int
}

func parseFlags() *config {
	c := &config{}

	flag.StringVar(&c.RendezvousString, "rendezvous", "meetme", "Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	flag.StringVar(&c.listenHost, "host", "0.0.0.0", "The bootstrap node host listen address\n")
	flag.StringVar(&c.ProtocolID, "pid", "/chat/1.1.0", "Sets a protocol id for stream headers")
	flag.StringVar(&c.bootstrap, "bootstrap", "", "set bootstrap")
	flag.StringVar(&c.id, "ID", "", "set name")
	flag.IntVar(&c.listenPort, "port", 4001, "node listen port")

	flag.Parse()
	return c
}
