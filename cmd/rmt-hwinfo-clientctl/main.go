// Derived from cmd/public-api-demo in github.com/SUSE/connect-ng's next branch
package main

import (
	"fmt"
	"log"

	"github.com/rtamalin/rmt-client-testing/internal/clientstore"
	"github.com/rtamalin/rmt-client-testing/internal/workqueue"
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

	wq := workqueue.NewWorkQueue("testq", cliOpts.NumJobs)

	wq.Start()
	for i := int64(0); i < cliOpts.NumClients; i++ {
		job := wq.NewJob(i, func() error {
			return performAction(uint32(i), &cliOpts)
		})
		wq.Add(job)
	}

	wq.WaitForCompletion()

	if len(wq.Errors) > 0 {
		log.Printf("ERROR: %v action failures occurred:\n", len(wq.Errors))
		for _, actErr := range wq.Errors {
			log.Printf("  %s\n", actErr.Error())
		}
		log.Fatal("ERROR: failed due to above errors.")
	}

	fmt.Println(wq.Stats.JobStats().Summary("Client"))
	fmt.Println(wq.Stats.PoolStats().Summary("Parallel Job"))
}
