// Derived from cmd/public-api-demo in github.com/SUSE/connect-ng's next branch
package main

import (
	"log"

	"github.com/rtamalin/rmt-client-testing/internal/clientstore"
	"github.com/rtamalin/rmt-client-testing/internal/flagtypes"
)

func performAction(id uint32, opts *CliOpts) (err error) {
	fileId := clientstore.FileId(id)
	switch opts.Action {
	case ACTION_REGISTER:
		err = registerClient(fileId, opts)
	case ACTION_UPDATE:
		err = updateClient(fileId, opts)
	case ACTION_DEREGISTER:
		err = deregisterClient(fileId, opts)
	}
	return
}

func main() {

	parseCliArgs(&cliOpts)

	for i := flagtypes.Uint32(0); i < cliOpts.NumClients; i++ {
		err := performAction(uint32(i), &cliOpts)

		if err != nil {
			log.Fatalf("Error: %s\n", err.Error())
		}
	}
}
