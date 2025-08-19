package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/rtamalin/rmt-client-testing/internal/client"
	"github.com/rtamalin/rmt-client-testing/internal/clientstore"
)

const (
	CREATE = true
)

type Uint32 uint32

func (u *Uint32) String() string {
	return strconv.FormatUint(uint64(*u), 10)
}

func (u *Uint32) Set(v string) (err error) {
	val, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		err = fmt.Errorf(
			"failed to parse %q as a Uint32: %w",
			v,
			err,
		)
		return
	}
	*u = Uint32(val)
	return
}

type Options struct {
	NumClients Uint32
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

	for i := Uint32(0); i < options.NumClients; i++ {
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
