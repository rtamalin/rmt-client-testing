// Derived from cmd/public-api-demo in github.com/SUSE/connect-ng's next branch
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

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

func SaveStats(opts *CliOpts, stats []string, stdout bool) (err error) {
	curTime := time.Now().UTC()

	// generate stats file content
	header := fmt.Sprintf(
		"[Start of client %s summary statistics at %s]",
		opts.Action.String(),
		curTime.Format(time.DateTime),
	)
	footer := "[End of summary statistics]"
	outputBlocks := []string{
		header,
	}
	outputBlocks = append(outputBlocks, stats...)
	outputBlocks = append(outputBlocks, footer)
	content := strings.Join(outputBlocks, "\n")

	if stdout {
		fmt.Println(content)
	}

	// generate stats file name based upon UTC timestamp
	statsFileName := fmt.Sprintf(
		"%s_%s_%s_%d.log",
		curTime.Format(time.DateOnly),
		strings.Replace(curTime.Format(time.TimeOnly), ":", "", -1),
		opts.Action.String(),
		opts.NumClients,
	)

	statsDir := filepath.Join(opts.DataStore, "stats")
	statsPath := filepath.Join(statsDir, statsFileName)

	// create the stats dir if needed
	err = os.MkdirAll(statsDir, 0o755)
	if err != nil {
		log.Printf(
			"ERROR: failed to create stats save dir %q: %s",
			statsDir,
			err.Error(),
		)
		return
	}

	// write stats content to stats file
	err = os.WriteFile(statsPath, []byte(content), 0o644)
	if err != nil {
		log.Printf(
			"ERROR: failed to create stats file %q: %s",
			statsPath,
			err.Error(),
		)
		return
	}

	return
}

func main() {

	parseCliArgs(&cliOpts)

	clientStatOpts := workqueue.SummaryOpts{
		workqueue.OPT_NAME:          "Client " + cliOpts.Action.String(),
		workqueue.OPT_RATE:          true,
		workqueue.OPT_MIN_MAX:       true,
		workqueue.OPT_EXTRA_STATS:   true,
		workqueue.OPT_DATA_PROFILES: !cliOpts.NoDataProfiles,
	}

	parallelStatOpts := workqueue.SummaryOpts{
		workqueue.OPT_NAME:        "Parallel Job",
		workqueue.OPT_MIN_MAX:     true,
		workqueue.OPT_EXTRA_STATS: true,
	}

	wq := workqueue.NewWorkQueue(cliOpts.Action.String(), cliOpts.NumJobs)

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

	SaveStats(
		&cliOpts,
		[]string{
			wq.Stats.JobStats().Summary(clientStatOpts),
			wq.Stats.PoolStats().Summary(parallelStatOpts),
		},
		true, /* write to stdout */
	)
}
