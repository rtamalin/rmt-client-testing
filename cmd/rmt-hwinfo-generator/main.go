package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/rtamalin/rmt-client-testing/internal/client"
	"github.com/rtamalin/rmt-client-testing/internal/clientstore"
	"github.com/rtamalin/rmt-client-testing/internal/flagtypes"
)

const (
	CREATE = true
)

type Options struct {
	NumClients flagtypes.Uint32
	DataStore  string
}

var option_defaults = Options{
	DataStore:  "ClientDataStore",
	NumClients: 1000,
}

var options Options

func main() {
	options = option_defaults
	flag.StringVar(&options.DataStore, "datastore", option_defaults.DataStore, "Location of `datastore` to store simulated clients")
	flag.Var(&options.NumClients, "clients", "The number of `clients` to simulate")
	flag.Parse()

	fmt.Printf("Initialising %q as datastore\n", options.DataStore)
	dataStore := clientstore.New(options.DataStore)

	fmt.Printf("Simulating %v clients\n", options.NumClients)

	for i := flagtypes.Uint32(0); i < options.NumClients; i++ {
		c := client.NewClient(client.ClientId(i))
		sysInfo := c.SystemInfo()

		func() {
			fp, err := dataStore.Open(clientstore.FileId(i), CREATE)
			if err != nil {
				log.Fatalf(
					"Failed to create client %v: %s",
					i,
					err.Error(),
				)
			}
			defer fp.Close()

			// truncate existing content
			err = fp.Truncate(0)
			if err != nil {
				log.Fatalf(
					"Failed to truncate client %v to %q: %s",
					i,
					fp.Name(),
					err.Error(),
				)
			}

			_, err = fp.Write([]byte(sysInfo))
			if err != nil {
				log.Fatalf(
					"Failed to write client %v to %q: %s",
					i,
					fp.Name(),
					err.Error(),
				)
			}
		}()

		//fmt.Printf("Client %v:\n%s\n\n", i, sysInfo)
	}
}
