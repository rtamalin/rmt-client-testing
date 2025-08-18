package main

import (
	"flag"
	"fmt"

	"github.com/rtamalin/rmt-client-testing/internal/client"
)

type Options struct {
	NumClients uint64
}

var option_defaults = Options{
	NumClients: 1000,
}

var options Options

func main() {
	flag.Uint64Var(&options.NumClients, "clients", option_defaults.NumClients, "The number of `clients` to simulate")
	flag.Parse()

	fmt.Printf("Simulating %v clients\n", options.NumClients)

	//clients := make([]*client.Client, options.NumClients)

	var i uint64
	for i = 0; i < options.NumClients; i++ {
		c := client.NewClient(client.ClientId(i))
		//clients[i] = c
		fmt.Printf("Client %v:\n%s\n\n", i, c.SystemInfo())
	}
}
